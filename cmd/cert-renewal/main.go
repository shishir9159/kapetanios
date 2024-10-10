package main

import (
	"bytes"
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

	// replace zap with zerolog

	c := Controller{
		ctx: context.Background(),
		log: logger,
	}

	GrpcClient(c.log)

	var outb, errb bytes.Buffer
	cmd := exec.Command("/bin/bash", "-c", "ls")
	cmd.Stdout = &outb
	cmd.Stderr = &errb

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

	GrpcClient(c.log)

	//	step 1. Backup directories
	err = BackupCertificatesKubeConfigs(c, backupCount)
	if err != nil {
		c.log.Error("failed to backup certificates and kubeConfigs",
			zap.Error(err))
	}

	//	step 2. Kubeadm certs renew all

	//    cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	err = cmd.Run()
	c.log.Info("ls after step 1 chroot",
		zap.String("output", outb.String()),
		zap.String("err", errb.String()))

	err = Renew(c)
	if err != nil {
		c.log.Error("failed to renew certificates and kubeConfigs",
			zap.Error(err))
	}

	err = cmd.Run()
	c.log.Info("ls -after renew step 2 chroot",
		zap.String("output", outb.String()),
		zap.String("err", errb.String()))

	//step 3. Restarting pods to work with the updated certificates
	err = Restart(c)
	if err != nil {
		c.log.Error("failed to restart kubernetes components after certificate renewal",
			zap.Error(err))
	}

	GrpcClient(c.log)
}
