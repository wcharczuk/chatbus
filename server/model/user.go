package model

import (
	"database/sql"
	"fmt"

	"github.com/blendlabs/spiffy"
)

// User is a connected user.
type User struct {
	ID          int    `json:"id" db:"id,pk,serial"`
	UUID        string `json:"uuid" db:"uuid"`
	DisplayName string `json:"display_name" db:"display_name"`
}

// IsZero returns if the user is set or not
func (u *User) IsZero() bool {
	return u.ID == 0
}

// TableName returns the table name for the object.
func (u User) TableName() string {
	return "users"
}

// Populate manually populates an object.
func (u *User) Populate(r *sql.Rows) error {
	return r.Scan(&u.ID, &u.UUID, &u.DisplayName)
}

// GetUserByUUID gets a user by uuid.
func GetUserByUUID(uuid string, txs ...*sql.Tx) (*User, error) {
	var tx *sql.Tx
	if len(txs) > 0 {
		tx = txs[0]
	}
	var user User
	queryBody := fmt.Sprintf("select %s from %s where uuid = $1", spiffy.ColumnNames(user), user.TableName())
	err := DB().QueryInTransaction(queryBody, tx, uuid).Out(&user)
	return &user, err
}
