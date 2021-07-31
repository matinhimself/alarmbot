package internal

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Reminder struct {
	Id primitive.ObjectID `bson:"_id,omitempty"`

	Description string `bson:"name"`

	AtTime time.Time     `bson:"time"`
	Every  time.Duration `bson:"every"`
	From   time.Time     `bson:"from"`

	Created    time.Time `bson:"created"`
	IsRepeated bool      `bson:"is_repeated"`
	Priority   Priority  `bson:"priority"`

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
	rem.IsRepeated = true
	return rem
}

func (r *Reminder) GetIdentifier() string {
	return r.Id.Hex()
}

func (r *Reminder) Validate() error {
	if r.AtTime.Before(r.Created) {
		return fmt.Errorf("time is not validated")
	}
	return nil
}

type Priority int8

func (p Priority) ToString() string {
	if p == Silent {
		return "unmute"
	} else {
		return "mute"
	}
}

func (p Priority) Not() Priority {
	if p == Silent {
		return Audible
	} else {
		return Silent
	}
}
