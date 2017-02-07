package controller

import (
	"time"

	chronometer "github.com/blendlabs/go-chronometer"
)

// CullSessions is the job that removes dead sessions
type CullSessions struct {
	Controller *Chat
}

// Name is the job name
func (cs CullSessions) Name() string {
	return "cull_sessions"
}

// Execute is the job body.
func (cs CullSessions) Execute(ct *chronometer.CancellationToken) error {
	ct.CheckCancellation()

	cutoff := time.Now().UTC().Add(-5 * time.Minute)
	var err error
	for _, session := range cs.Controller.Sessions {
		if session.LastActiveUTC.Before(cutoff) {
			err = cs.Controller.deleteSession(session)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Schedule returns the job schedule.
func (cs CullSessions) Schedule() chronometer.Schedule {
	return chronometer.EveryMinute()
}
