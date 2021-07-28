package internal

type User struct {
	UserId   int    `bson:"_id"`
	Username string `bson:"username"`
	Name     string `bson:"name"`
}
