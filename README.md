chatbus
=======

chat bus is a microservice backend for chat applications. it is designed to be accessed using an interval polling method, i.e. every N seconds you request what messages arrived for a session
in the past N seconds. 

tips:
- the server will restore its state when restarting from the db automatically, warming the caches.
- due to sensitivity around timing, it is important you use server timestamps to track timing. do not use client timestamps as the client machines clock may be skewed. 
- this service is almost totally unauthenticated; authentication should be handled upstream of this service.
- sessions are marked 'last_active' when they check for messages and when they send messages. there is a job that culls messages older than 5 minutes.
- the timestamp arguments to `/api/messages/:session_id/:unix/:nano` are structured such that the `:unix` argument should be a unix timestamp, i.e. seconds since epoch, and `:nano` should be any remainder in nanoseconds.
- the underlying message queue implementation is a ringbuffer that we seek into in reverse order. source can be found [here](https://github.com/blendlabs/go-util/blob/master/collections/ring_buffer.go)

## prerequisites

- go 1.6+
- postgres 9.5+

## getting started

- clone this repo to `$GOPATH/src/github.com/blendlabs/chatbus`
- run `make db`
- run `make run`
- profit!

## performance

is decent but not mindblowing. the service is very careful to maintain synchronization with the cache layers, and uses deferred writes and a couple other strategies to keep message throughput high.

an example of local performance when fetching messages with both unix and unix nano timestamp cutoffs: 
```
wrk -t 16 -c 16 http://127.0.0.1:8080/api/messages/97e0a981300f41d7b2f092596d7dff3c/1471756932/632474775
Running 10s test @ http://127.0.0.1:8080/api/messages/97e0a981300f41d7b2f092596d7dff3c/1471756932/632474775
  16 threads and 16 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    16.23ms   14.21ms 132.63ms   86.18%
    Req/Sec    70.78     27.45   171.00     64.35%
  11335 requests in 10.04s, 1.97GB read
Requests/sec:   1128.89
Transfer/sec:    200.95MB
``` 
considering we're moving about 2.0gb over the wire this is not bad. 