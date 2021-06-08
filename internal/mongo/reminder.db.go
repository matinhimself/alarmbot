package mongo

import (
	"context"
	"fmt"
	"github.com/psyg1k/remindertelbot/internal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func (db *Db) DeleteRemindersBefore(id int64, t time.Time) (int64, error) {
	collection := db.RemindersCollection

	filter := bson.M{
		"$and": []bson.M{
			{"chat_id": id},
			{"time": bson.M{
				"$lt": primitive.NewDateTimeFromTime(
					t,
				),
			},
			},
		},
	}

	dr, err := collection.DeleteMany(nil, filter)

	if err != nil {
		return 0, err
	}

	return dr.DeletedCount, err
}

func (db *Db) DeleteReminder(reminderID int64) error {
	collection := db.RemindersCollection

	filter := bson.D{{"_id", reminderID}}

	_, err := collection.DeleteOne(nil, filter)
	return err
}

func (db *Db) GetReminder(id int64) (internal.Reminder, error) {
	objId, _ := primitive.ObjectIDFromHex(fmt.Sprintf("%x", id))

	var reminder internal.Reminder

	err := db.RemindersCollection.FindOne(
		context.TODO(),
		bson.M{
			"name": objId,
		},
	).Decode(&reminder)

	if err != nil {
		return reminder, err
	}

	return reminder, nil
}

func (db *Db) GetRemindersByChatID(chatId int64) (reminders []internal.Reminder, err error) {

	reminders = make([]internal.Reminder, 0)

	collection := db.RemindersCollection

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

	collection := db.RemindersCollection

	// Mongo go driver doesn't support auto-increasing unique ids
	// so I implement it manually.
	// it stores last saved id in database and increases it everytime
	// gets next sequence number
	id, err := db.getNextSeq("id")
	if err != nil {
		return internal.Reminder{}, err
	}

	r.Id = id

	_, err = collection.InsertOne(nil, r)
	return r, err
}

type count struct {
	Id   primitive.ObjectID `bson:"Id"`
	Name string             `bson:"name"`
	Seq  int64              `bson:"seq"`
}

func (db *Db) getNextSeq(name string) (int64, error) {
	counter := db.CounterCollection

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

	return c.Seq, err
}

func (db *Db) UpdatePriority(id int64, priority int8) error {

	collection := db.RemindersCollection

	filter := bson.D{{"_id", id}}

	update := bson.D{
		{"$set", bson.D{
			{"priority", priority},
		}},
	}

	_, err := collection.UpdateOne(context.TODO(), filter, update)
	return err
}
