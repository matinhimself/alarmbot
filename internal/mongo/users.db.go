package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

import "github.com/psyg1k/remindertelbot/internal"

func (db *Db) InsertUser(user internal.User) error {
	collection := db.usersCollection

	_, err := collection.InsertOne(nil, user)

	return err
}

func (db *Db) UpdateUserTz(location time.Location, userID int) error {
	collection := db.usersCollection

	filter := bson.D{{"_id", userID}}

	update := bson.D{
		{"$set", bson.D{
			{"offset", location},
		}},
	}

	_, err := collection.UpdateOne(nil, filter, update)
	return err
}

func (db *Db) UpdateLanguage(lang internal.Language, chatId int64) error {
	collection := db.chatsCollection

	filter := bson.D{{"_id", chatId}}

	update := bson.D{
		{"$set", bson.D{
			{"lang", lang},
		}},
	}

	_, err := collection.UpdateOne(nil, filter, update)
	return err
}

func (db *Db) GetUser(id int) (u internal.User, err error) {
	collection := db.usersCollection

	filter := bson.D{{"_id", id}}
	err = collection.FindOne(nil, filter).Decode(&u)
	if err != nil {
		return u, err
	}

	return u, nil
}
