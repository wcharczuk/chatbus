package model

import "database/sql"

// Contacts is a contact list entry
type Contacts struct {
	Sender   int `json:"sender" db:"sender,pk"`
	Receiver int `json:"receiver" db:"receiver,pk"`
}

// IsZero returns if the object is set or not
func (c Contacts) IsZero() bool {
	return c.Sender == 0 || c.Receiver == 0
}

// TableName returns the table name for the object.
func (c Contacts) TableName() string {
	return "contacts"
}

// DeleteContacts deletes a contacts pair for a sender and receiver.
func DeleteContacts(sender, receiver int, txs ...*sql.Tx) error {
	var tx *sql.Tx
	if len(txs) > 0 {
		tx = txs[0]
	}

	queryBody := `
	DELETE FROM contacts s
	where 
		(sender = $1 and receiver = $2) 
		or (sender = $2 and receiver = $1)
	`

	return DB().ExecInTransaction(queryBody, tx, sender, receiver)
}
