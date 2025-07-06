package model

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Mongo struct {
	Dsn    string
	Client *mongo.Client
}

func (m *Mongo) InitConnection() error {
	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://root:5uZLvWUrvJ@kcprod-mongodb-svc.lighthouse.svc.cluster.local:27017"))
	if err != nil {
		panic(err)
	}

	m.Client = client
	return nil
}

func (m *Mongo) CloseConnection() error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := m.client.Disconnect(ctx); err != nil {
		return err
	}

	return nil
}

func (m *Mongo) MongoDB() {

	//
	//_ = client.Ping(ctx, readpref.Primary())
	//
	//collection := client.Database("testing").Collection("numbers")
	//
	//res, err := collection.InsertOne(ctx, bson.D{{"name", "pi"}, {"value", 3.14159}})
	//if err != nil {
	//	panic(err)
	//}
	//
	//id := res.InsertedID
	//
	//fmt.Println(id)
}
