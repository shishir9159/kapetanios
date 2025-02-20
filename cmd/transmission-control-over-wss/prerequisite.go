package main

import (
	"go.uber.org/zap"
)

// if the labels exists
// if the labels match with the expectations
// number of control-plane pods and nodes running
// if the nodes match with the expectations

// TODO: controller needs to be passed

// TODO: node - role name

func Prerequisites(upgrade *Upgrade) {

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

	//configMapName := "kubeadm-config"
	//configMap, er := Upgrade.nefario.client.Clientset().CoreV1().ConfigMaps("kube-system").
	//	Get(context.Background(), configMapName, metav1.GetOptions{})
	//if er != nil {
	//	Upgrade.nefario.log.Info("error fetching the kubeadm-config configMap: ",
	//		zap.Error(er))
	//}

	if report.NodesToBeUpgraded == "" {
		return
	}

	upgrade.mu.Lock()
	defer upgrade.mu.Unlock()

	// TODO: race condition - readCtx can be cancelled
	upgrade.nefario.log.Info("upgrade is continued after the successful restart",
		zap.String("nodes to be upgraded", report.NodesToBeUpgraded))

	upgrade.upgraded = make(chan bool, 1)

	// TODO: another channel to send signal
	//  that server can run now that mutex is locked
	upgrade.MinorUpgrade(report)
}
