package mongo

import (
	"context"
	"crypto/rand"
	"github.com/oklog/ulid/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
	. "kapetanios/internal/model"
	"time"
)

type Mongo struct {
	reportID   string
	DSN        string
	DB         string
	Collection string
	Log        *zap.Logger
	Client     *mongo.Client
}

func (m *Mongo) InitConnection() error {

	client, err := mongo.Connect(options.Client().ApplyURI(m.DSN))
	if err != nil {
		// TODO:
		//  look for the error when mongo is containerCreating
		//  panic(err)
		m.Log.Error("MongoDB connection error",
			zap.Error(err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = client.Ping(ctx, nil); err != nil {
		panic(err)
	}

	m.Client = client

	dbs, err := m.Client.ListDatabaseNames(ctx, nil)
	// TODO:
	//  if error equals to document is nil
	if err != nil {
		m.Log.Error("mongodb list database names",
			zap.Error(err))
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
			//TODO:
			// Collection already exists
			//panic(err)
			m.Log.Error("mongodb list database names",
				zap.Error(err))
		}
	}

	exists = false
	for _, db := range dbs {
		if db == "report" {
			exists = true
			break
		}
	}

	if !exists {
		err = m.Client.Database(m.DB).CreateCollection(ctx, "report")
		if err != nil {
			//TODO:
			// Collection already exists
			//panic(err)
			m.Log.Error("mongodb list database names",
				zap.Error(err))
		}
	}

	collections, err := m.Client.Database(m.DB).ListCollectionNames(ctx, bson.M{"name": m.Collection})
	// TODO:
	//  if error equals to collection is nil
	if err != nil {
		m.Log.Error("mongodb list collection names", zap.Error(err))
	}

	if len(collections) != 1 {
		// TODO:
		// panic(fmt.Errorf("collections not found"))

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

// TODO:
//  refactor logic

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

func (m *Mongo) MinorUpgradeReport() error {

	entropy := ulid.Monotonic(rand.Reader, 0)
	id := ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
	m.reportID = id.String()

	collection := m.Client.Database("lighthouse").Collection("report")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, Report{
		ID: m.reportID,
	})
	if err != nil {
		m.Log.Warn("failed to read config", zap.Error(err))
		return err
	}

	// TODO:
	m.Log.Info("updated config", zap.Any("result", result))

	return nil
}

func (m *Mongo) MinorUpgradeReportUpdate(nodeReport NodeReport) error {

	collection := m.Client.Database("lighthouse").Collection("report")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var report Report
	entry := bson.M{"_id": m.reportID}
	err := collection.FindOne(ctx, entry).Decode(&report)
	if err != nil {
		m.Log.Warn("failed to read config", zap.Error(err))
	}

	report.NodesReport = append(report.NodesReport, nodeReport)
	result, err := collection.UpdateOne(ctx, entry, report)
	if err != nil {
		m.Log.Warn("failed to read config", zap.Error(err))
		return err
	}

	// TODO:
	m.Log.Info("updated config", zap.Any("result", result))

	return nil
}
