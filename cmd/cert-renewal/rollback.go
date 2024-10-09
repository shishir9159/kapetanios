package main

import (
	"log"
)

// kubelet error scraping

func getKubeConfigFiles() []string {

	// could be extracted from the static manifest files
	// --kubeconfig=/etc/kubernetes/controller-manager.conf

	kubeConfigFiles := []string{
		"admin.conf",
		"controller-manager.conf",
	}

	return kubeConfigFiles
}

func OverRide() {

}

func Rollout() error {

	kubeConfigs := getKubeConfigFiles()
	dest := "/etc/kubernetes"

	for _, kubeConfigFile := range kubeConfigs {
		err := Copy(kubeConfigFile, dest)
		if err != nil {
			//fmt.Println(backupDir, kubernetesConfigDir, kubeConfigs)
			log.Println(err)
		}
	}

	return nil
}
