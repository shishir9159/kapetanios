package main

import (
	"context"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
)

// if the labels exists
// if the labels match with the expectations
// number of control-plane pods and nodes running
// if the nodes match with the expectations

// TODO: node - role name

func Prerequisites(upgrade *MinorUpgrade) error {

	// if cm shows updated nodes to a certain value
	// and desired kubernetesVersion version must exist on the cm
	// for the updates

	// TODO: throw error no master nodes found

	//configMapName := "kubeadm-config"
	//configMap, er := Upgrade.nefario.client.Clientset().CoreV1().ConfigMaps("kube-system").
	//	Get(context.Background(), configMapName, metav1.GetOptions{})
	//if er != nil {
	//	Upgrade.nefario.log.Info("error fetching the kubeadm-config configMap: ",
	//		zap.Error(er))
	//}

	// TODO:
	// 	clean up pods in internal library

	//labelSelector := metav1.LabelSelector{
	//	MatchLabels: map[string]string{
	//		"app": "minor-upgrade",
	//		"app": "certs",
	//	},
	//}

	// TODO:
	//  3 mintues timeout for the deletion to complete

	listOptions := metav1.ListOptions{
		LabelSelector: "app in (certs, minor-upgrade), app notin (etcd, kubelet, control-plane)",
		//LabelSelector: "app in (certs, minor-upgrade), app notin (etcd, kubelet, control-plane)",
		//LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	minions, err := upgrade.nefario.client.Clientset().CoreV1().Pods(os.Getenv("KAPETANIOS_NAMESPACE")).
		List(context.Background(), listOptions)
	if err != nil {
		upgrade.nefario.log.Error("error listing completed/failed pods for node management",
			zap.Error(err))
		return err
	}

	for _, minion := range minions.Items {
		er := upgrade.nefario.client.Clientset().CoreV1().Pods(os.Getenv("KAPETANIOS_NAMESPACE")).
			Delete(context.Background(), minion.Name, metav1.DeleteOptions{})
		if er != nil {
			upgrade.nefario.log.Info("failed to delete minion:",
				zap.String("minion name:", minion.Name),
				zap.Error(er))
		}
	}

	// TODO: race condition - readCtx can be cancelled
	upgrade.nefario.log.Info("upgrade is continued after the successful restart",
		zap.Strings("nodes to be upgraded", upgrade.upgradeState.NodesStatus.NodesToBeUpgraded))

	upgrade.finalizer = make(chan bool, 1)

	upgrade.mu.Lock()
	go upgrade.MinorUpgrade()

	return nil
}
