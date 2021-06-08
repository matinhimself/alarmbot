package mongo

import (
	ctx "context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Db struct {
	client              *mongo.Client
	RemindersCollection *mongo.Collection
	UsersCollection     *mongo.Collection
	ChatsCollection     *mongo.Collection

	// a collection for storing last id stored in database
	// TODO: migration
	CounterCollection *mongo.Collection
}

func NewDb(URI string) (*Db, error) {
	client, err := mongo.Connect(ctx.TODO(), options.Client().ApplyURI(URI))
	return &Db{
		client:              client,
		RemindersCollection: client.Database("reminder_bot").Collection("bot"),
		UsersCollection:     client.Database("reminder_bot").Collection("users"),
		ChatsCollection:     client.Database("reminder_bot").Collection("chats"),
		CounterCollection:   client.Database("reminder_bot").Collection("counter"),
	}, err
}

func (db *Db) Ping() error {
	return db.client.Ping(ctx.TODO(), nil)
}
