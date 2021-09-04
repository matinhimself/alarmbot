package mongo

import (
	"context"
	"github.com/psyg1k/remindertelbot/internal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func (db *Db) DeleteRemindersBefore(id int64, t time.Time) (int64, error) {
	collection := db.remindersCollection

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

	dr, err := collection.DeleteMany(context.TODO(), filter)

	if err != nil {
		return 0, err
	}

	return dr.DeletedCount, err
}

func (db *Db) DeleteReminder(idh string) error {
	id, err := primitive.ObjectIDFromHex(idh)
	if err != nil {
		return err
	}
	collection := db.remindersCollection

	filter := bson.D{primitive.E{Key: "_id", Value: id}}

	_, err = collection.DeleteOne(context.TODO(), filter)
	return err
}

func (db *Db) GetReminder(idh string) (internal.Reminder, error) {
	id, _ := primitive.ObjectIDFromHex(idh)

	var reminder internal.Reminder

	err := db.remindersCollection.FindOne(
		context.TODO(),
		bson.M{
			"name": id,
		},
	).Decode(&reminder)

	if err != nil {
		return reminder, err
	}

	return reminder, nil
}

func (db *Db) GetRemindersByChatID(chatId int64) (reminders []internal.Reminder, err error) {

	reminders = make([]internal.Reminder, 0)

	collection := db.remindersCollection
	// Sort by
	findOptions := options.Find().SetSort(bson.D{{Key: "time", Value: 1}})

	filter := bson.D{primitive.E{Key: "chat_id", Value: chatId}}

	cur, err := collection.Find(context.TODO(), filter, findOptions)

	if err != nil {
		return reminders, err
	}

	for cur.Next(context.TODO()) {
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

	collection := db.remindersCollection

	one, err := collection.InsertOne(context.TODO(), r)
	if err != nil {
		return internal.Reminder{}, err
	}
	oid, _ := one.InsertedID.(primitive.ObjectID)
	r.Id = oid
	return r, nil

}

func (db *Db) UpdatePriority(idh string, priority internal.Priority) error {

	collection := db.remindersCollection
	id, err := primitive.ObjectIDFromHex(idh)
	if err != nil {
		return err
	}
	filter := bson.D{primitive.E{Key: "_id", Value: id}}

	update := bson.D{
		{"$set", bson.D{
			{"priority", priority},
		}},
	}

	_, err = collection.UpdateOne(context.TODO(), filter, update)
	return err
}
