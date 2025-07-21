package mongo

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
	. "kapetanios/internal/model"
	"time"
)

type Mongo struct {
	minorUpgradeID bson.ObjectID
	DSN            string
	DB             string
	Collection     string
	Log            *zap.Logger
	Client         *mongo.Client
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
			// panic(err)
			m.Log.Error("mongodb list database names",
				zap.Error(err))
		}
	}

	exists = false
	for _, db := range dbs {
		if db == "history" {
			exists = true
			break
		}
	}

	if !exists {
		err = m.Client.Database(m.DB).CreateCollection(ctx, "history")
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
		if db == "certExpirationReport" {
			exists = true
			break
		}
	}

	if !exists {
		err = m.Client.Database(m.DB).CreateCollection(ctx, "certExpirationReport")
		if err != nil {
			//TODO:
			// Collection already exists
			// panic(err)
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
		//  panic(fmt.Errorf("collections not found"))

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

// Last Config is not useful for mongo,
// but configMap is good

func (m *Mongo) ReadConfig() (UpgradeState, error) {

	collection := m.Client.Database("lighthouse").Collection("minorUpgrade")

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

	collection := m.Client.Database("lighthouse").Collection("minorUpgrade")

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

// TODO:
//  update the _id value after a restart

func (m *Mongo) MinorUpgradeReport() error {

	collection := m.Client.Database("lighthouse").Collection("history")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, ClusterMinorUpgradeHistory{})
	if err != nil {
		m.Log.Warn("failed to read config", zap.Error(err))
		return err
	}

	if oid, ok := result.InsertedID.(bson.ObjectID); ok {
		m.minorUpgradeID = oid
		return nil
	} else {
		return errors.New("failed to read object id")
	}
}

func (m *Mongo) MinorUpgradeReportUpdate(nodeReport NodeMinorUpgradeHistory) error {

	collection := m.Client.Database("lighthouse").Collection("history")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var report ClusterMinorUpgradeHistory
	entry := bson.M{"_id": m.minorUpgradeID}
	err := collection.FindOne(ctx, entry).Decode(&report)
	if err != nil {
		m.Log.Warn("failed to read config", zap.Error(err))
	}

	report.NodesMinorUpgradeHistory = append(report.NodesMinorUpgradeHistory, nodeReport)
	_, err = collection.ReplaceOne(ctx, entry, report)
	if err != nil {
		m.Log.Warn("failed to read config", zap.Error(err))
		return err
	}

	return nil
}

func (m *Mongo) LastMinorUpgradeReport() (ClusterMinorUpgradeHistory, error) {

	collection := m.Client.Database("lighthouse").Collection("history")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var report ClusterMinorUpgradeHistory

	//findOne := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}})
	//curr, err := collection.Find(ctx, bson.D{}, findOne)
	//
	//var reportList []Report
	//for curr.Next(ctx) {
	//	elem := reflect.New(reflect.TypeOf(Report{})).Interface()
	//	err = curr.Decode(elem)
	//
	//	if err != nil {
	//		m.Log.Warn("failed to read config", zap.Error(err))
	//	}
	//
	//	elem3 := elem.(*Report)
	//	m.Log.Info("updated config", zap.Any("result", elem3))
	//	reportList = append(reportList, *elem3)
	//}
	//
	//m.Log.Info("length of report",
	//	zap.Int("result", len(reportList)))

	opts := options.FindOne().SetSort(bson.D{{Key: "_id", Value: -1}})
	err := collection.FindOne(ctx, bson.M{}, opts).Decode(&report)

	if errors.Is(err, mongo.ErrNoDocuments) {
		m.Log.Warn("mongo.ErrNoDocuments",
			zap.Error(err))
		return ClusterMinorUpgradeHistory{}, err
	} else if err != nil {
		m.Log.Warn("failed to read report",
			zap.Error(err))
		return ClusterMinorUpgradeHistory{}, err
	}

	return report, nil
}

//func (m *Mongo) LastMinorUpgradeReport() (Report, error) {
//
//	collection := m.Client.Database("lighthouse").Collection("history")
//
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//
//	findOne := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}})
//	curr, err := collection.Find(ctx, bson.D{}, findOne)
//
//	var reportList []Report
//	for curr.Next(ctx) {
//		elem := reflect.New(reflect.TypeOf(Report{})).Interface()
//		err = curr.Decode(elem)
//
//		if err != nil {
//			m.Log.Warn("failed to read config", zap.Error(err))
//		}
//
//		elem3 := elem.(*Report)
//		m.Log.Info("updated config", zap.Any("result", elem3))
//		reportList = append(reportList, *elem3)
//	}
//
//	m.Log.Info("length of report",
//		zap.Int("result", len(reportList)))
//
//	return reportList[0], nil
//}

//	result := collection.FindOne(ctx, opts)
//  obj := reflect.New(reflect.TypeOf(Report{})).Interface()
//  err := result.Decode(obj)
//  if err != nil {
//    m.Log.Warn("failed to read config",
//    zap.Error(err))
//  }
//
//  report = *obj.(*Report)

// TODO:
//  abstract under common behavior pattern

func (m *Mongo) CertificateReportUpdate(k8sCertsExpiration KubernetesCertificatesExpiration) error {

	collection := m.Client.Database("lighthouse").Collection("certExpirationReport")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, k8sCertsExpiration)
	if err != nil {
		m.Log.Warn("failed to read config", zap.Error(err))
		return err
	}

	return nil
}

func (m *Mongo) LastCertificateReport() (KubernetesCertificatesExpiration, error) {

	collection := m.Client.Database("lighthouse").Collection("certExpirationReport")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var report KubernetesCertificatesExpiration

	opts := options.FindOne().SetSort(bson.D{{Key: "_id", Value: -1}})
	err := collection.FindOne(ctx, bson.M{}, opts).Decode(&report)

	if errors.Is(err, mongo.ErrNoDocuments) {
		m.Log.Warn("mongo.ErrNoDocuments",
			zap.Error(err))
		return KubernetesCertificatesExpiration{}, err
	} else if err != nil {
		m.Log.Warn("failed to read report",
			zap.Error(err))
		return KubernetesCertificatesExpiration{}, err
	}

	return report, nil
}
