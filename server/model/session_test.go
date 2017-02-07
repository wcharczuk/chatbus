package model

import (
	"testing"
	"time"

	assert "github.com/blendlabs/go-assert"
	util "github.com/blendlabs/go-util"
)

func TestSessionCreate(t *testing.T) {
	assert := assert.New(t)
	tx, err := DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u1 := &User{UUID: util.UUIDv4().ToShortString(), DisplayName: "Test User"}
	assert.Nil(DB().CreateInTransaction(u1, tx))
	s1 := &Session{
		UUID:          util.UUIDv4().ToShortString(),
		CreatedUTC:    time.Now().UTC(),
		LastActiveUTC: time.Now().UTC(),
		UserID:        u1.ID,
		User:          u1,
	}
	assert.Nil(DB().CreateInTransaction(s1, tx))
}
