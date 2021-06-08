package internal

type User struct {
	UserId   int      `bson:"_id"`
	Username string   `bson:"username"`
	Name     string   `bson:"name"`
	Language Language `bson:"language"`
}

const (
	ENGLISH = "en"
	FARSI   = "fa"
)

type Language string

func (l Language) ToString() string {
	if l == ENGLISH {
		return "English"
	} else if l == FARSI {
		return "Farsi/فارسی"
	}
	return "undetected"
}
