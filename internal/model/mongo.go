package model

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

func MongoDB() {

	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_ = client.Ping(ctx, readpref.Primary())

	collection := client.Database("testing").Collection("numbers")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, _ := collection.InsertOne(ctx, bson.D{{"name", "pi"}, {"value", 3.14159}})
	id := res.InsertedID

}
