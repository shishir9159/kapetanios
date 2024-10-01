package main

import (
	"log"
	"os/exec"
	"syscall"
	"time"
)

///usr/local/bin/kubeadm certs renew

func Renew() error {
	err := syscall.Chroot("/host")
	if err != nil {
		//log.Println("Failed to create chroot on /host\n\n\n")
		log.Println(err)
	}

	// whereis kubeadm
	//"/usr/local/bin/kubeadm certs renew scheduler.conf"
	// it is assumed that kubeadm exist otherwise, cert validity wouldn't have work

	cmd := exec.Command("/usr/bin/kubeadm", "certs", "renew", "all", "--config=kubeadm-config.yaml")

	//    cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Println(err)
		time.Sleep(10 * time.Second)
	}

	return nil
}
