package main

import (
	"log"
)

func main() {

	//	step 1. Backup directories
	err := BackupCertificatesKubeConfigs(7)
	if err != nil {
		log.Println(err)
	}

	//	step 2. Kubeadm certs renew all
	err = Renew()
	if err != nil {
		log.Println(err)
	}

	//step 3. Restarting pods to work with the updated certificates
	err = Restart()
	if err != nil {
		log.Println(err)
	}
}
