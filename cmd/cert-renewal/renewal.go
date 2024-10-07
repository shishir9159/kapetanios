package main

import (
	"go.uber.org/zap"
	"log"
	"os/exec"
	"syscall"
	"time"
)

///usr/local/bin/kubeadm certs renew

func Renew(c Controller) error {
	err := syscall.Chroot("/host")
	if err != nil {
		c.log.Error("Failed to create chroot on /host", zap.Error(err))
	}

	// whereis kubeadm
	//"/usr/local/bin/kubeadm certs renew scheduler.conf"
	// it is assumed that kubeadm exist otherwise, cert validity wouldn't have work

	cmd := exec.Command("/usr/bin/kubeadm", "certs", "renew", "all", "--config=kubeadm-config.yaml")

	//    cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Println(err)
		c.log.Error("Failed to renew certificates", zap.Error(err))
		time.Sleep(10 * time.Second)
	}

	return nil
}
