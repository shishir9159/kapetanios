package model

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type Mongo struct {
	dsn  string
	conn *mongo.Client
}

func (m *Mongo) InitConnection() error {

	return nil
}

func (m *Mongo) MongoDB() {
	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://root:5uZLvWUrvJ@kcprod-mongodb-svc.lighthouse.svc.cluster.local:27017"))
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	_ = client.Ping(ctx, readpref.Primary())

	collection := client.Database("testing").Collection("numbers")

	res, err := collection.InsertOne(ctx, bson.D{{"name", "pi"}, {"value", 3.14159}})
	if err != nil {
		panic(err)
	}

	id := res.InsertedID

	fmt.Println(id)
}
