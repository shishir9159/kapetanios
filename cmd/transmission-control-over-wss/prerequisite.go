package main

import (
	"context"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// if the labels exists
// if the labels match with the expectations
// number of control-plane pods and nodes running
// if the nodes match with the expectations

// TODO: controller needs to be passed

// TODO: node - role name

func Prerequisites(upgrade *upgrade) {

	// if cm shows updated nodes to a certain value
	// and desired kubernetesVersion version must exist on the cm
	// for the updates

	// TODO: throw error no master nodes found

	report, err := readConfig(upgrade.nefario)
	if err != nil {
		upgrade.nefario.log.Info("failed to read config map: ",
			zap.Error(err))
		return
	}

	// TODO: controller should be created and passed here

	configMapName := "kubeadm-config"

	//configMap, er := upgrade.nefario.client.Clientset().CoreV1().ConfigMaps("kube-system").
	//	Get(context.Background(), configMapName, metav1.GetOptions{})
	//if er != nil {
	//	upgrade.nefario.log.Info("error fetching the kubeadm-config configMap: ",
	//		zap.Error(er))
	//}

}
