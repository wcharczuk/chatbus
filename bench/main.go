package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/blendlabs/chatbus/server/model"
	request "github.com/blendlabs/go-request"
	util "github.com/blendlabs/go-util"
)

type serviceResponseOfUser struct {
	Meta     map[string]interface{} `json:"meta"`
	Response model.User             `json:"response"`
}

type serviceResponseOfSession struct {
	Meta     map[string]interface{} `json:"meta"`
	Response model.Session          `json:"response"`
}

func createUser(index int) (int, error) {
	var res serviceResponseOfUser
	err := request.NewHTTPRequest().
		AsPost().
		WithHost("localhost:8080").
		WithPathf("/api/users").
		WithJSONBody(model.User{UUID: fmt.Sprintf("test_user_%d", index)}).
		FetchJSONToObject(&res)

	if err != nil {
		return 0, err
	}

	return res.Response.ID, nil
}

func createSession(userID int) (string, error) {
	var res serviceResponseOfUser
	err := request.NewHTTPRequest().
		AsPost().
		WithHost("localhost:8080").
		WithPathf("/api/session/%d", userID).
		FetchJSONToObject(&res)

	if err != nil {
		return "", err
	}

	return res.Response.UUID, nil
}

func getRoot() *request.HTTPRequest {
	return request.NewHTTPRequest().
		AsPost().
		WithHost("127.0.0.1:8080").
		WithPathf("/").
		WithKeepAlives().WithTimeout(5 * time.Second)
}

func sendMessage(sessionID string, receiverID int) *request.HTTPRequest {
	return request.NewHTTPRequest().
		AsPost().
		WithHost("127.0.0.1:8080").
		WithPathf("/api/send/%s", sessionID).
		WithJSONBody(model.Message{ReceiverID: receiverID, Body: util.UUIDv4().ToShortString()}).
		WithKeepAlives().WithTimeout(5 * time.Second)
}

func getMessages(sessionID string, cutoff time.Time) *request.HTTPRequest {
	unix := cutoff.Unix()
	nano := cutoff.UnixNano() - int64(time.Duration(cutoff.Unix())*(time.Second/time.Nanosecond))
	return request.NewHTTPRequest().
		AsGet().
		WithHost("127.0.0.1:8080").
		WithPathf("/api/messages/%s/%d/%d", sessionID, unix, nano).
		WithKeepAlives().WithTimeout(5 * time.Second)
}

func benchmark(threads, connections int, duration time.Duration, benchmarkName string, requestFactory func() *request.HTTPRequest) {
	fmt.Printf("Running %s Benchmark ...", benchmarkName)

	var requestCount int32
	var errorCount int32
	var bytesRead int32
	benchmarkStart := time.Now()

	requests := make([]*request.HTTPRequest, connections)
	for x := 0; x < connections; x++ {
		requests[x] = requestFactory()
	}

	buffers := make([]*bytes.Buffer, connections)
	for x := 0; x < connections; x++ {
		buffers[x] = bytes.NewBuffer([]byte{})
	}

	transports := make([]*http.Transport, connections)

	wg := sync.WaitGroup{}
	wg.Add(threads)

	defer func() {
		benchmarkElapsed := time.Now().Sub(benchmarkStart)
		throughput := float64(requestCount) / (float64(benchmarkElapsed) / float64(time.Second))
		fmt.Println(" Complete!")
		fmt.Printf("Time Elapsed: %v, Request Count: %d, Error Count: %d, Throughput: %0.2f rps Bytes Read: %s\n", benchmarkElapsed, requestCount, errorCount, throughput, util.FormatFileSize(int(bytesRead)))
	}()

	connectionsPerThread := connections / threads

	for x := 0; x < threads; x++ {
		go func(threadID int) {
			defer wg.Done()
			var err error
			var requestIndex int
			for time.Now().Sub(benchmarkStart) < duration {
				for y := 0; y < connectionsPerThread; y++ {
					requestIndex = y + (threadID * connectionsPerThread)
					err = requests[requestIndex].WithTransport(transports[requestIndex]).OnCreateTransport(func(_ *url.URL, t *http.Transport) {
						transports[requestIndex] = t
					}).WithResponseBuffer(buffers[requestIndex]).Execute()
					if err != nil {
						atomic.AddInt32(&errorCount, 1)
					}
					atomic.AddInt32(&requestCount, 1)
					atomic.AddInt32(&bytesRead, int32(buffers[requestIndex].Len()))
					buffers[requestIndex].Reset()
				}
			}
		}(x)
	}
	wg.Wait()
}

func main() {
	threads := flag.Int("threads", runtime.NumCPU(), "number of goroutines (a proxy for threads) to use.")
	connections := flag.Int("connections", runtime.NumCPU()*8, "total number of connections to maintain.")
	rawDuration := flag.String("duration", "15s", "test duration")
	flag.Parse()

	duration, err := time.ParseDuration(*rawDuration)
	if err != nil {
		log.Fatal(err)
	}

	uid1, err := createUser(1)
	if err != nil {
		log.Fatal(err)
	}
	uid2, err := createUser(2)
	if err != nil {
		log.Fatal(err)
	}

	sid1, err := createSession(uid1)
	if err != nil {
		log.Fatal(err)
	}

	/*benchmark("Root", func() *request.HTTPRequest {
		return getRoot()
	})*/

	fmt.Printf("Benchmark Settings threads:%d connections:%d duration:%v\n\n", *threads, *connections, duration)
	benchmark(*threads, *connections, duration, "Send Message", func() *request.HTTPRequest {
		return sendMessage(sid1, uid2)
	})

	println()

	benchmark(*threads, *connections, duration, "Get Messages", func() *request.HTTPRequest {
		return getMessages(sid1, time.Now().Add(-10*time.Millisecond))
	})
}
