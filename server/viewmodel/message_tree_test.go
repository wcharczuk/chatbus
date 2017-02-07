package viewmodel

import (
	"fmt"
	"testing"
	"time"

	"github.com/blendlabs/chatbus/server/model"
	assert "github.com/blendlabs/go-assert"
)

func TestMessageTree(t *testing.T) {
	assert := assert.New(t)

	bt := NewTreeOfMessage()

	now := time.Now().UTC()

	var messages []model.Message
	for x := 0; x < 32; x++ {
		messages = append(messages, model.Message{
			UUID:       fmt.Sprintf("m%d", x),
			CreatedUTC: now.Add(time.Duration(-x) * time.Second),
		})
	}
	for _, m := range messages {
		bt.Set(m.CreatedUTC, m)
	}

	cursor, _ := bt.Seek(now.Add(-16 * time.Second).Add(-1 * time.Microsecond))
	assert.NotNil(cursor)

	var foundMessages []string
	_, v, err := cursor.Next()
	for err == nil {
		message := model.TryCastMessage(v)
		foundMessages = append(foundMessages, message.UUID)

		_, v, err = cursor.Next()
	}
	assert.Len(foundMessages, 17)
	assert.Equal("m16", foundMessages[0])
	assert.Equal("m15", foundMessages[1])
	assert.Equal("m14", foundMessages[2])

	k, _ := bt.Last()
	assert.Equal(messages[0].CreatedUTC, k)

	ok := bt.Delete(k)
	assert.True(ok)
	assert.Equal(31, bt.Len())
}
