package model

import (
	"testing"

	assert "github.com/blendlabs/go-assert"
	util "github.com/blendlabs/go-util"
)

func TestContactsCreate(t *testing.T) {
	assert := assert.New(t)
	tx, err := DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u1 := &User{UUID: util.UUIDv4().ToShortString(), DisplayName: "Test User"}
	assert.Nil(DB().CreateInTransaction(u1, tx))

	u2 := &User{UUID: util.UUIDv4().ToShortString(), DisplayName: "Test User 2"}
	assert.Nil(DB().CreateInTransaction(u2, tx))

	contact := &Contacts{Sender: u1.ID, Receiver: u2.ID}
	assert.Nil(DB().CreateInTransaction(contact, tx))
}
