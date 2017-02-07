package model

import (
	"strings"

	"github.com/blendlabs/spiffy"
)

// EscapeJSON escapes quotes.
func EscapeJSON(input string) string {
	return strings.Replace(input, `"`, `\"`, -1)
}

// DB is the current db.
func DB() *spiffy.DbConnection {
	return spiffy.DefaultDb()
}
