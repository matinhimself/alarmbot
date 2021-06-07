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

	Created time.Time `bson:"created"`

	Priority int8 `bson:"priority"`

	ChatId  int64 `bson:"chat_id"`
	Message int   `bson:"description"`
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
