package mongo

import (
	"context"
	"fmt"
	"github.com/psyg1k/remindertelbot/internal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

func (db *Db) GetReminder(id int64) internal.Reminder {
	objId, _ := primitive.ObjectIDFromHex(fmt.Sprintf("%x", id))

	var reminder internal.Reminder

	err := db.RemindersCollection.FindOne(context.TODO(), bson.M{"name": objId}).Decode(&reminder)
	if err != nil {
		log.Printf("method: Get, Id: %x, %s", id, err)
	}
	return reminder
}

func (db *Db) GetRemindersByChatID(chatId int64) (reminders []internal.Reminder, err error) {

	reminders = make([]internal.Reminder, 0)

	collection := db.client.Database("jobs").Collection("reminders")

	// Sort by
	findOptions := options.Find().SetSort(bson.D{{"time", 1}})

	filter := bson.D{{"chat_id", chatId}}

	cur, err := collection.Find(nil, filter, findOptions)

	if err != nil {
		return reminders, err
	}

	for cur.Next(nil) {
		var reminder internal.Reminder

		err := cur.Decode(&reminder)
		if err != nil {
			return reminders, err
		}

		reminders = append(reminders, reminder)
	}
	return reminders, nil
}

func (db *Db) InsertReminder(r internal.Reminder) (internal.Reminder, error) {

	collection := db.client.Database("jobs").Collection("reminders")

	// Mongo go driver doesn't support auto-increasing unique ids
	// so I implemented it manually
	// it stores last saved id in database and increases it everytime
	// gets next sequence number
	r.Id = db.getNextSeq("id")

	_, err := collection.InsertOne(nil, r)
	return r, err
}

func (db *Db) getNextSeq(name string) int64 {
	counter := db.client.Database("jobs").Collection("counter")

	var c count
	err := counter.FindOneAndUpdate(
		nil,
		bson.M{"name": name},
		bson.M{
			"$inc": bson.M{
				"seq": 1,
			},
		},
	).Decode(&c)
	if err != nil {
		log.Println(err)
	}
	return c.Seq
}

type count struct {
	Id   primitive.ObjectID `bson:"Id"`
	Name string             `bson:"name"`
	Seq  int64              `bson:"seq"`
}
