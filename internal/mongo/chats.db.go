package mongo

import (
	"fmt"
	"github.com/psyg1k/remindertelbot/internal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (db *Db) DeleteTaskList(chatId int64) error {
	collection := db.chatsCollection

	filter := bson.D{{"_id", chatId}}

	_, err := collection.DeleteOne(nil, filter)
	return err
}

func (db *Db) GetChat(chatId int64) (chat internal.Chat, err error) {
	collection := db.chatsCollection

	filter := bson.D{{"_id", chatId}}

	err = collection.FindOne(nil, filter).Decode(&chat)

	if err != nil {
		return internal.Chat{}, err
	}
	return
}

func (db *Db) GetTaskListMessageId(chatId int64) (int, error) {
	collection := db.chatsCollection

	var chat internal.Chat

	filter := bson.D{{"_id", chatId}}

	err := collection.FindOne(nil, filter).Decode(&chat)

	if err == mongo.ErrNoDocuments {
		return 0, fmt.Errorf("tasklist not registered")
	} else if err != nil {
		return 0, err
	}

	return chat.TaskList, nil
}

func (db *Db) InsertChat(chat internal.Chat) error {
	collection := db.chatsCollection

	_, err := collection.InsertOne(nil, chat)

	return err
}
