package model

import (
	"time"

	util "github.com/blendlabs/go-util"
	"github.com/blendlabs/spiffy"
)

// NewSession creates a new session
func NewSession(userID int) (*Session, error) {
	var user User
	err := spiffy.DefaultDb().GetByID(&user, userID)
	if err != nil {
		return nil, err
	}
	return &Session{
		UUID:       util.UUIDv4().ToShortString(),
		CreatedUTC: time.Now().UTC(),
		UserID:     userID,
		User:       &user,
	}, nil
}

// Session is a session for a user.
type Session struct {
	UUID          string    `json:"uuid" db:"uuid,pk"`
	CreatedUTC    time.Time `json:"created_utc" db:"created_utc"`
	LastActiveUTC time.Time `json:"last_active_utc" db:"last_active_utc"`
	UserID        int       `json:"user_id" db:"user_id"`
	User          *User     `json:"user" db:"-"`
}

// IsZero returns if the session is set or not.
func (s Session) IsZero() bool {
	return len(s.UUID) == 0
}

// TableName returns the table name.
func (s Session) TableName() string {
	return "sessions"
}
