package model

import (
	"testing"
	"time"

	assert "github.com/blendlabs/go-assert"
	util "github.com/blendlabs/go-util"
)

func TestGetAllMessagesWithLimit(t *testing.T) {
	assert := assert.New(t)
	tx, err := DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u1 := &User{UUID: util.UUIDv4().ToShortString(), DisplayName: "Test User"}
	assert.Nil(DB().CreateInTransaction(u1, tx))

	u2 := &User{UUID: util.UUIDv4().ToShortString(), DisplayName: "Test User 2"}
	assert.Nil(DB().CreateInTransaction(u2, tx))

	u3 := &User{UUID: util.UUIDv4().ToShortString(), DisplayName: "Test User 2"}
	assert.Nil(DB().CreateInTransaction(u3, tx))

	assert.Nil(DB().CreateInTransaction(&Message{UUID: util.UUIDv4().ToShortString(), CreatedUTC: time.Now().UTC(), SenderID: u1.ID, ReceiverID: u2.ID, Body: "Test"}, tx))
	assert.Nil(DB().CreateInTransaction(&Message{UUID: util.UUIDv4().ToShortString(), CreatedUTC: time.Now().UTC(), SenderID: u1.ID, ReceiverID: u2.ID, Body: "Test"}, tx))
	assert.Nil(DB().CreateInTransaction(&Message{UUID: util.UUIDv4().ToShortString(), CreatedUTC: time.Now().UTC(), SenderID: u1.ID, ReceiverID: u2.ID, Body: "Test"}, tx))
	assert.Nil(DB().CreateInTransaction(&Message{UUID: util.UUIDv4().ToShortString(), CreatedUTC: time.Now().UTC(), SenderID: u1.ID, ReceiverID: u2.ID, Body: "Test"}, tx))
	assert.Nil(DB().CreateInTransaction(&Message{UUID: util.UUIDv4().ToShortString(), CreatedUTC: time.Now().UTC(), SenderID: u1.ID, ReceiverID: u2.ID, Body: "Test"}, tx))
	assert.Nil(DB().CreateInTransaction(&Message{UUID: util.UUIDv4().ToShortString(), CreatedUTC: time.Now().UTC(), SenderID: u1.ID, ReceiverID: u2.ID, Body: "Test"}, tx))
	assert.Nil(DB().CreateInTransaction(&Message{UUID: util.UUIDv4().ToShortString(), CreatedUTC: time.Now().UTC(), SenderID: u1.ID, ReceiverID: u2.ID, Body: "Test"}, tx))

	assert.Nil(DB().CreateInTransaction(&Message{UUID: util.UUIDv4().ToShortString(), CreatedUTC: time.Now().UTC(), SenderID: u1.ID, ReceiverID: u3.ID, Body: "Test"}, tx))
	assert.Nil(DB().CreateInTransaction(&Message{UUID: util.UUIDv4().ToShortString(), CreatedUTC: time.Now().UTC(), SenderID: u1.ID, ReceiverID: u3.ID, Body: "Test"}, tx))
	assert.Nil(DB().CreateInTransaction(&Message{UUID: util.UUIDv4().ToShortString(), CreatedUTC: time.Now().UTC(), SenderID: u1.ID, ReceiverID: u3.ID, Body: "Test"}, tx))

	messages, err := GetAllMessagesWithLimit(5, tx)

	var filtered []Message
	for _, m := range messages {
		if m.SenderID == u1.ID {
			filtered = append(filtered, m)
		}
	}

	assert.Nil(err)

	assert.Len(filtered, 8)
}
