package mongo

import "go.mongodb.org/mongo-driver/bson"

import "github.com/psyg1k/remindertelbot/internal"

func (db *Db) InsertUser(username string, userId int, name string, lang internal.Language) error {
	collection := db.UsersCollection

	if lang == "" {
		lang = internal.ENGLISH
	}

	_, err := collection.InsertOne(nil, internal.User{
		UserId:   userId,
		Username: username,
		Name:     name,
		Language: lang,
	})

	return err
}

func (db *Db) UpdateLanguage(lang string, chatId int64) error {
	collection := db.ChatsCollection

	filter := bson.D{{"_id", chatId}}

	update := bson.D{
		{"$set", bson.D{
			{"lang", lang},
		}},
	}

	_, err := collection.UpdateOne(nil, filter, update)
	return err
}

func (db *Db) GetUsers(id int) (users []*internal.User, err error) {
	collection := db.UsersCollection

	filter := bson.D{{"_id", id}}

	cur, err := collection.Find(nil, filter)
	if err != nil {
		return users, err
	}

	for cur.Next(nil) {
		var user internal.User

		err := cur.Decode(&user)
		if err != nil {
			continue
		}

		users = append(users, &user)
	}
	return users, nil
}
