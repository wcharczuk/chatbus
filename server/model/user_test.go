package model

import (
	"testing"

	assert "github.com/blendlabs/go-assert"
	util "github.com/blendlabs/go-util"
)

func TestUserCreate(t *testing.T) {
	assert := assert.New(t)
	tx, err := DB().Begin()
	assert.Nil(err)
	defer tx.Rollback()

	u1 := &User{UUID: util.UUIDv4().ToShortString(), DisplayName: "Test User"}
	assert.Nil(DB().CreateInTransaction(u1, tx))
}
