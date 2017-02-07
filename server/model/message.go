package model

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/blendlabs/go-workqueue"
	"github.com/blendlabs/spiffy"
)

// TryCastMessage tries to cast an interface as a *Message
func TryCastMessage(obj interface{}) *Message {
	if typed, isTyped := obj.(Message); isTyped {
		return &typed
	}
	if typed, isTyped := obj.(*Message); isTyped {
		return typed
	}
	return nil
}

// Message is a sent message
type Message struct {
	UUID        string                 `json:"uuid" db:"uuid,pk"`
	CreatedUTC  time.Time              `json:"created_utc" db:"created_utc"`
	SenderID    int                    `json:"sender_id" db:"sender"`
	ReceiverID  int                    `json:"receiver_id" db:"receiver"` // REQUIRED
	Sender      *User                  `json:"sender,omitempty" db:"-"`
	Receiver    *User                  `json:"receiver,omitempty" db:"-"`
	Body        string                 `json:"body" db:"body"` // REQUIRED (MAYBE??)
	Attachments map[string]interface{} `json:"attachments" db:"attachments,json"`
}

// IsZero returns if the object is set or not.
func (m Message) IsZero() bool {
	return len(m.UUID) == 0
}

// LessThan returns if an object
func (m Message) LessThan(other interface{}) bool {
	if typed, isTyped := other.(Message); isTyped {
		return m.CreatedUTC.Before(typed.CreatedUTC)
	}
	if typed, isTyped := other.(*Message); isTyped {
		return m.CreatedUTC.Before(typed.CreatedUTC)
	}
	return false
}

// TableName returns the table name for the object.
func (m Message) TableName() string {
	return "messages"
}

// QueueCreate queue's a message create.
func (m Message) QueueCreate() {
	workQueue.Enqueue(func(v ...interface{}) error {
		if len(v) == 0 {
			return nil
		}
		if typed, isTyped := v[0].(spiffy.DatabaseMapped); isTyped {
			return DB().Create(typed)
		}
		return nil
	}, m)
}

// GetAllMessagesWithLimit gets all the messages within a given limit (per recipient).
func GetAllMessagesWithLimit(limit int, txs ...*sql.Tx) ([]Message, error) {
	var tx *sql.Tx
	if len(txs) > 0 {
		tx = txs[0]
	}
	var messages []Message

	queryFormat := `
	SELECT %s FROM
	(
		SELECT
			ROW_NUMBER() over (PARTITION BY m.receiver ORDER BY m.created_utc desc) as rank
			, m.*
		FROM 
			%s m
	) as datums
	where datums.rank <= $1
	order by datums.created_utc asc
	`
	queryBody := fmt.Sprintf(queryFormat, spiffy.ColumnNames(Message{}), Message{}.TableName())
	err := DB().QueryInTransaction(queryBody, tx, limit).OutMany(&messages)
	return messages, err
}
