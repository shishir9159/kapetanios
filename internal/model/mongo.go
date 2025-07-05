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

func MongoDB() {
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

	res, _ := collection.InsertOne(ctx, bson.D{{"name", "pi"}, {"value", 3.14159}})
	id := res.InsertedID

	fmt.Println(id)
}
