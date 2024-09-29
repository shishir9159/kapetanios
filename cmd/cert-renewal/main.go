package cert_renewal

import "log"

func main() {
	//	step 1. Backup directories
	err := BackupCertificatesKubeConfigs(3)
	if err != nil {
		log.Println(err)
	}
}
