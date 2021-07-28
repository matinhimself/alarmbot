package internal

import (
	"fmt"
	"time"
)

type Reminder struct {
	Id int64 `bson:"_id"`

	Description string `bson:"name"`

	AtTime time.Time     `bson:"time"`
	Every  time.Duration `bson:"every"`
	From   time.Time     `bson:"from"`

	Created    time.Time `bson:"created"`
	IsRepeated bool      `bson:"is_repeated"`
	Priority   int8      `bson:"priority"`

	ChatId  int64 `bson:"chat_id"`
	Message int   `bson:"description"`
}

const (
	Silent  = 0
	Audible = 1
)

const MaxTimeUnix = 1 << 34 // "github.com/jalaali/go-jalaali" for reasons
// doesn't support whole int64 as a unix time :( so for max available time
// I should use 1<<35 instead of 1<<63

var MaxTime = time.Unix(MaxTimeUnix, 0)

func WithRepeat(rem *Reminder) *Reminder {
	rem.AtTime = time.Unix(MaxTimeUnix<<1, 0)
	return rem
}

func (r *Reminder) GetIdentifier() int64 {
	return r.Id
}

func (r *Reminder) Validate() error {
	if r.AtTime.Before(r.Created) {
		return fmt.Errorf("time is not validated")
	}
	return nil
}

const (
	// Muted, alarms just in last tick
	PriorityLow Priority = iota

	PriorityHigh
)

type Priority int8

func (p Priority) Validate() error {
	switch p {
	case PriorityLow, PriorityHigh:
		return nil
	}

	return fmt.Errorf("priority validation failed")
}
