package internal

import "time"

type Chat struct {
	ChatID    int64    `bson:"_id"`
	TaskList  int      `bson:"task_list"`
	UTCOffset Offset   `bson:"offset"`
	IsJalali  bool     `bson:"is_jalali"`
	Language  Language `bson:"lang"`
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
