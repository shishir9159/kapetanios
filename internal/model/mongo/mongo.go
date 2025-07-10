package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
	. "kapetanios/internal/model"
	"time"
)

type Mongo struct {
	DSN        string
	DB         string
	Collection string
	Log        *zap.Logger
	Client     *mongo.Client
}

func (m *Mongo) InitConnection() error {

	client, err := mongo.Connect(options.Client().ApplyURI(m.DSN))
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = client.Ping(ctx, nil); err != nil {
		panic(err)
	}

	m.Client = client

	dbs, err := m.Client.ListDatabaseNames(ctx, nil)
	if err != nil {
		panic(err)
	}

	exists := false
	for _, db := range dbs {
		if db == m.DB {
			exists = true
			break
		}
	}

	if !exists {
		err = m.Client.Database(m.DB).CreateCollection(ctx, m.Collection)
		if err != nil {
			panic(err)
		}
	}

	collections, err := m.Client.Database(m.DB).ListCollectionNames(ctx, bson.M{"name": m.Collection})
	if err != nil {
		panic(err)
	}

	if len(collections) != 1 {

		//panic(fmt.Errorf("collections not found"))

		err = m.Client.Database(m.DB).CreateCollection(ctx, m.Collection)
		if err != nil {
			panic(err)
		}
	}

	return nil
}

func (m *Mongo) CloseConnection() error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := m.Client.Disconnect(ctx); err != nil {
		return err
	}

	return nil
}

func (m *Mongo) ReadConfig() (UpgradeState, error) {

	collection := m.Client.Database("lighthouse").Collection("minor-upgrade")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.FindOne().SetSort(bson.D{{Key: "_id", Value: -1}})
	var state UpgradeState
	err := collection.FindOne(ctx, UpgradeState{}, opts).Decode(&state)
	if err != nil {
		m.Log.Warn("failed to read config", zap.Error(err))
		return UpgradeState{}, err
	}

	return state, nil
}

func (m *Mongo) WriteConfig(state UpgradeState) error {

	collection := m.Client.Database("lighthouse").Collection("minor-upgrade")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, state)
	if err != nil {
		m.Log.Warn("failed to read config", zap.Error(err))
		return err
	}

	// TODO:
	m.Log.Info("updated config", zap.Any("result", result))

	return nil
}
