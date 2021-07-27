package internal

import (
	"time"
)

type User struct {
	UserId   int      `bson:"_id"`
	Username string   `bson:"username"`
	Name     string   `bson:"name"`
	Language Language `bson:"language"`
	Offset   Offset   `bson:"offset"`
}

const (
	ENGLISH = "en"
	FARSI   = "fa"
)

const (
	IRAN = -1 * 210 * time.Minute
)

var tzs = [...]Offset{
	Offset(IRAN),
}

var Languages = [...]Language{
	Language(FARSI),
	Language(ENGLISH),
}

type Language string
type Offset time.Duration

func (l Language) ToString() string {
	if l == ENGLISH {
		return "English"
	} else if l == FARSI {
		return "Farsi/فارسی"
	}
	return "undetected"
}
