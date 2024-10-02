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

	//	step 2.
	err = Renew()
	if err != nil {
		log.Println(err)
	}

	//step 3.
	err = Restart()
	if err != nil {
		log.Println(err)
	}
}
