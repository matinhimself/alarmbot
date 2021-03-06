package mongo

import (
	ctx "context"
	"github.com/psyg1k/remindertelbot/internal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
	"log"
	"time"
)

type Db struct {
	client              *mongo.Client
	remindersCollection *mongo.Collection
	usersCollection     *mongo.Collection
	chatsCollection     *mongo.Collection

	// a collection for storing last id stored in database
	// TODO: migration
	CounterCollection *mongo.Collection
}

func NewDb(URI string) (*Db, error) {
	client, err := mongo.Connect(ctx.TODO(), options.Client().ApplyURI(URI))
	return &Db{
		client:              client,
		remindersCollection: client.Database("reminder_bot").Collection("bot"),
		usersCollection:     client.Database("reminder_bot").Collection("users"),
		chatsCollection:     client.Database("reminder_bot").Collection("chats"),
		CounterCollection:   client.Database("reminder_bot").Collection("counter"),
	}, err
}

func (db *Db) Ping() error {
	return db.client.Ping(ctx.TODO(), nil)
}

func (db *Db) GetRemindersAfter(now time.Time) ([]internal.Reminder, error) {
	var reminders []internal.Reminder

	filter := bson.M{
		"time": bson.M{"$gt": primitive.NewDateTimeFromTime(now)},
	}

	cur, err := db.remindersCollection.Find(context.TODO(), filter)
	if err != nil {
		log.Println(err)
		return reminders, err
	}

	for cur.Next(context.TODO()) {
		var reminder internal.Reminder
		err := cur.Decode(&reminder)
		if err != nil {
			continue
		}
		reminders = append(reminders, reminder)
	}

	return reminders, nil
}

func (db *Db) GetRemindersAfterOrIsRepeated(now time.Time) ([]internal.Reminder, error) {
	var reminders []internal.Reminder

	filter := bson.M{
		"$or": []bson.M{
			{"time": bson.M{
				"$gt": primitive.NewDateTimeFromTime(now),
			},
			},
			{
				"is_repeated": bson.M{
					"$eq": true,
				},
			},
		},
	}

	cur, err := db.remindersCollection.Find(context.TODO(), filter)
	if err != nil {
		log.Println(err)
		return reminders, err
	}

	for cur.Next(context.TODO()) {
		var reminder internal.Reminder
		err := cur.Decode(&reminder)
		if err != nil {
			continue
		}
		reminders = append(reminders, reminder)
	}

	return reminders, nil
}
