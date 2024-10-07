package main

import (
	"go.uber.org/zap"
)

func main() {

	logger, err := zap.NewProduction()
	defer func(logger *zap.Logger) {
		er := logger.Sync()
		if er != nil {

		}
	}(logger)

	//zap.ReplaceGlobals(logger)

	//logger.Info("failed to fetch URL",
	//	// Structured context as strongly typed Field values.
	//	//zap.String("url", url),
	//	zap.Int("attempt", 3),
	//	zap.Duration("backoff", time.Second),
	//)

	//client, err := orchestration.NewClient()
	//if err != nil {
	//	logger.Fatal("error creating Kubernetes client: ",
	//		zap.Error(err))
	//}

	//	step 1. Backup directories
	err = BackupCertificatesKubeConfigs(7)
	if err != nil {
		logger.Error("failed to backup certificates and kubeConfigs",
			zap.Error(err))
	}

	//	step 2. Kubeadm certs renew all
	err = Renew(logger)
	if err != nil {
		logger.Error("failed to renew certificates and kubeConfigs",
			zap.Error(err))
	}

	//step 3. Restarting pods to work with the updated certificates
	err = Restart()
	if err != nil {
		logger.Error("failed to restart kubernetes components after certificate renewal",
			zap.Error(err))
	}
}
