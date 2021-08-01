package mongo

import (
	"context"
	"fmt"
	"github.com/psyg1k/remindertelbot/internal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (db *Db) DeleteTaskList(id int64) error {
	collection := db.chatsCollection

	filter := bson.D{primitive.E{Key: "_id", Value: id}}

	_, err := collection.DeleteOne(context.TODO(), filter)
	return err
}

func (db *Db) GetChat(id int64) (chat internal.Chat, err error) {
	collection := db.chatsCollection

	filter := bson.D{primitive.E{Key: "_id", Value: id}}

	err = collection.FindOne(context.TODO(), filter).Decode(&chat)

	if err != nil {
		return internal.Chat{}, err
	}
	return
}

func (db *Db) GetTaskListMessageId(id int64) (int, error) {
	collection := db.chatsCollection

	var chat internal.Chat

	filter := bson.D{primitive.E{Key: "_id", Value: id}}

	err := collection.FindOne(context.TODO(), filter).Decode(&chat)

	if err == mongo.ErrNoDocuments {
		return 0, fmt.Errorf("tasklist not registered")
	} else if err != nil {
		return 0, err
	}

	return chat.TaskList, nil
}

func (db *Db) InsertChat(chat internal.Chat) error {
	collection := db.chatsCollection

	_, err := collection.InsertOne(context.TODO(), chat)

	return err
}

func (db *Db) UpdateChatTz(id int64, location string) error {

	collection := db.chatsCollection

	filter := bson.D{primitive.E{Key: "_id", Value: id}}

	update := bson.D{
		primitive.E{
			Key: "$set",
			Value: primitive.E{
				Key:   "loc",
				Value: location,
			},
		},
	}

	_, err := collection.UpdateOne(context.TODO(), filter, update)
	return err
}

func (db *Db) UpdateChatCal(id int64, isJalali bool) error {

	collection := db.chatsCollection

	filter := bson.D{primitive.E{Key: "_id", Value: id}}

	update := bson.D{
		primitive.E{
			Key: "$set",
			Value: primitive.E{
				Key:   "is_jalali",
				Value: isJalali,
			},
		},
	}

	_, err := collection.UpdateOne(context.TODO(), filter, update)
	return err
}

func (db *Db) UpdateChatTaskList(id int64, messageId int) error {

	collection := db.chatsCollection

	filter := bson.D{primitive.E{Key: "_id", Value: id}}

	update := bson.D{
		primitive.E{
			Key: "$set",
			Value: primitive.E{
				Key:   "task_list",
				Value: messageId,
			},
		},
	}

	_, err := collection.UpdateOne(context.TODO(), filter, update)
	return err
}
