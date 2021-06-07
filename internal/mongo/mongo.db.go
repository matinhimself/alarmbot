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
}

func NewDb(URI string) (*Db, error) {
	client, err := mongo.Connect(ctx.TODO(), options.Client().ApplyURI(URI))
	return &Db{
		client:              client,
		RemindersCollection: client.Database("jobs").Collection("bot"),
		UsersCollection:     nil,
		ChatsCollection:     nil,
	}, err
}

func (db *Db) Ping() error {
	return db.client.Ping(ctx.TODO(), nil)
}
