package internal

type Chat struct {
	ChatID   int64    `bson:"_id"`
	TaskList int      `bson:"task_list"`
	Loc      string   `bson:"loc"`
	IsJalali bool     `bson:"is_jalali"`
	Language Language `bson:"lang"`
}

const (
	ENGLISH = "en"
	FARSI   = "fa"
)

const (
	IRAN = "Asia/Tehran"
)

const (
	GeoCal   = "gregorian"
	HijriCal = "hijri"
)

var Languages = [...]Language{
	Language(FARSI),
	Language(ENGLISH),
}

type Language string

func (l Language) ToString() string {
	if l == ENGLISH {
		return "English"
	} else if l == FARSI {
		return "Farsi/فارسی"
	}
	return "undetected"
}
