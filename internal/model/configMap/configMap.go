package configMap

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	. "kapetanios/internal/model"
)

type ConfigMap struct {
	Name      string
	NameSpace string
}

func (cm *ConfigMap) InitConnection() {
	// TODO: check if configMap exists
}

func (cm *ConfigMap) readConfig(nefario *Nefario) (UpgradeState, error) {
	configMap, err := nefario.Client.Clientset().CoreV1().ConfigMaps(nefario.Namespace).
		Get(nefario.Ctx, cm.Name, metav1.GetOptions{})
	if err != nil {
		nefario.Log.Error("error fetching the configMap",
			zap.Error(err))
		return UpgradeState{}, err
	}

	state := UpgradeState{}

	err = json.Unmarshal([]byte(configMap.Data["UPGRADE_STATE"]), &state)
	if err != nil {
		nefario.Log.Error("error unmarshalling the configMap",
			zap.Error(err))
	}

	if state.NodesStatus == nil {
		state.NodesStatus = &NodesStatus{
			NodesToBeUpgraded: []string{},
		}
	}

	return state, nil
}

func (cm *ConfigMap) writeConfig(nefario *Nefario, state UpgradeState) error {
	configMap, er := nefario.Client.Clientset().CoreV1().ConfigMaps(nefario.Namespace).
		Get(context.Background(), cm.Name, metav1.GetOptions{})
	if er != nil {
		nefario.Log.Error("error fetching the configMap",
			zap.Error(er))
	}

	payload, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		nefario.Log.Error("error marshaling the configMap",
			zap.Error(err))
	}

	configMap.Data["UPGRADE_STATE"] = string(payload)

	_, er = nefario.Client.Clientset().CoreV1().ConfigMaps(nefario.Namespace).
		Update(context.TODO(), configMap, metav1.UpdateOptions{})
	if er != nil {
		nefario.Log.Error("error updating configMap",
			zap.Error(er))
	}

	return nil
}
