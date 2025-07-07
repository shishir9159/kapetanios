package model

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type Mongo struct {
	Dsn    string
	Client *mongo.Client
}

func (m *Mongo) InitConnection() error {
	client, err := mongo.Connect(options.Client().ApplyURI(m.Dsn))
	if err != nil {
		panic(err)
	}

	m.Client = client
	return nil
}

func (m *Mongo) readConfig() {

}

func (m *Mongo) CloseConnection() error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := m.Client.Disconnect(ctx); err != nil {
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

func readConfig(nefario *Nefario) (UpgradeState, error) {

	configMapName := "kapetanios"

	configMap, err := nefario.client.Clientset().CoreV1().ConfigMaps(nefario.namespace).
		Get(nefario.ctx, configMapName, metav1.GetOptions{})
	if err != nil {
		nefario.log.Error("error fetching the configMap",
			zap.Error(err))
		return UpgradeState{}, err
	}

	state := UpgradeState{}

	err = json.Unmarshal([]byte(configMap.Data["UPGRADE_STATE"]), &state)
	if err != nil {
		nefario.log.Error("error unmarshalling the configMap",
			zap.Error(err))
	}

	if state.NodesStatus == nil {
		state.NodesStatus = &nodesStatus{
			NodesToBeUpgraded: []string{},
		}
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
