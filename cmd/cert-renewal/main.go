package main

import (
	"context"
	"go.uber.org/zap"
	"os/exec"
)

var (
	backupCount = 7
)

type Controller struct {
	ctx context.Context
	log *zap.Logger
}

func main() {

	logger, err := zap.NewProduction()
	defer func(logger *zap.Logger) {
		er := logger.Sync()
		if er != nil {

		}
	}(logger)

	c := Controller{
		ctx: context.Background(),
		log: logger,
	}

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
	err = BackupCertificatesKubeConfigs(c, backupCount)
	if err != nil {
		c.log.Error("failed to backup certificates and kubeConfigs",
			zap.Error(err))
	}

	//	step 2. Kubeadm certs renew all

	cmd := exec.Command("ls", "-la")

	//    cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	err = cmd.Run()
	if err != nil {
		c.log.Error("Failed to list directories",
			zap.Error(err))
	}

	err = Renew(c)
	if err != nil {
		c.log.Error("failed to renew certificates and kubeConfigs",
			zap.Error(err))
	}

	//step 3. Restarting pods to work with the updated certificates
	err = Restart(c)
	if err != nil {
		c.log.Error("failed to restart kubernetes components after certificate renewal",
			zap.Error(err))
	}

	GrpcClient(c.log)
}
