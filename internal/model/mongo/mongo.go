package mongo

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	. "kapetanios/internal/model"
	"time"
)

type Mongo struct {
	DSN        string
	DB         string
	Collection string
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
		panic(fmt.Errorf("collections not found"))
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

func (m *Mongo) readConfig(nefario *Nefario) (UpgradeState, error) {

	collection := m.Client.Database("lighthouse").Collection("minor-upgrade")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.FindOne().SetSort(bson.D{{Key: "_id", Value: -1}})
	var result bson.M
	err := collection.FindOne(ctx, bson.D{}, opts).Decode(&result)
	if err != nil {
		return UpgradeState{}, err
	}

	return state, nil
}

func writeConfig(nefario *Nefario, state UpgradeState) error {

	configMapName := "kapetanios"

	configMap, er := nefario.client.Clientset().CoreV1().ConfigMaps(nefario.namespace).
		Get(context.Background(), configMapName, metav1.GetOptions{})
	if er != nil {
		nefario.log.Error("error fetching the configMap",
			zap.Error(er))
	}

	payload, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		nefario.log.Error("error marshaling the configMap",
			zap.Error(err))
	}

	configMap.Data["UPGRADE_STATE"] = string(payload)

	_, er = nefario.client.Clientset().CoreV1().ConfigMaps(nefario.namespace).
		Update(context.TODO(), configMap, metav1.UpdateOptions{})
	if er != nil {
		nefario.log.Error("error updating configMap",
			zap.Error(er))
	}

	return nil
}
