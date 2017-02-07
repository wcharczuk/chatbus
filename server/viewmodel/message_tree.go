package viewmodel

import (
	"time"

	"github.com/blendlabs/chatbus/server/model"
	"github.com/cznic/b"
)

func messageCompare(v, k interface{}) int {
	vt := model.TryCastMessage(v)
	kt := model.TryCastMessage(k)
	if vt != nil && kt != nil {
		delta := vt.CreatedUTC.Sub(kt.CreatedUTC)
		return int(delta / time.Microsecond)
	}
	return 0
}

func timestampCompare(v, k interface{}) int {
	vt, isTyped := v.(time.Time)
	if !isTyped {
		return 0
	}
	kt, isTyped := k.(time.Time)
	if !isTyped {
		return 0
	}
	delta := vt.Sub(kt)
	return int(delta / time.Microsecond)
}

// NewTreeOfMessage returns a new message tree.
func NewTreeOfMessage() *b.Tree {
	return b.TreeNew(timestampCompare)
}
