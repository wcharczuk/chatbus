package viewmodel

import "github.com/blendlabs/chatbus/server/model"

// Contact is a contact list item.
type Contact struct {
	User     *model.User `json:"user"`
	IsOnline bool        `json:"is_online"`
	IsTyping bool        `json:"is_typing"`
}
