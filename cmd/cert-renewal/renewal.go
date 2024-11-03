package main

import (
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"go.uber.org/zap"
	"os/exec"
	"time"
)

//  kubeadm kubeconfig --help
//  export KUBECONFIG=$HOME/.kube/config

//  kubeadm kubeconfig user --org system:nodes --client-name system:node:$(hostname) --config=/etc/kubernetes/kubeadm/kubeadm-config.yaml > /etc/kubernetes/kubelet.conf.new

///usr/local/bin/kubeadm certs renew

func Renew(c Controller, connection pb.RenewalClient) error {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Error("Failed to create chroot on /host",
			zap.Error(err))
		return err
	}

	//  cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	// whereis kubeadm
	// "/usr/local/bin/kubeadm certs renew scheduler.conf"
	// it is assumed that kubeadm exist otherwise, cert validity wouldn't have work

	cmd := exec.Command("/usr/bin/kubeadm", "certs", "renew", "all", "--config=/etc/kubernetes/kubeadm-config.yaml")

	//    cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	err = cmd.Run()
	if err != nil {
		c.log.Error("Failed to renew certificates",
			zap.Error(err))
		time.Sleep(3 * time.Second)
	}

	if err = changedRoot(); err != nil {
		c.log.Fatal("Failed to exit from the updated root",
			zap.Error(err))
	}

	rpc, err := connection.BackupUpdate(c.ctx,
		&pb.BackupStatus{
			EtcdBackup:              false,
			KubeConfigBackup:        false,
			FileChecklistValidation: false,
			Err:                     "",
		})

	rpc, err := connection.Re

	if err != nil {
		c.log.Error("could not send status update: ", zap.Error(err))
	}

	c.log.Info("Backup Status",
		zap.Bool("next step", rpc.GetProceedNextStep()),
		zap.Bool("retry", rpc.GetSkipRetryCurrentStep()),
		zap.Bool("terminate application", rpc.GetTerminateApplication()))

	return err
}
