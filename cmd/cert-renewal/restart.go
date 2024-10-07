package main

import (
	"fmt"
	"log"
	"os/exec"
	"syscall"
	"time"
)

func restartService(component string) error {

	// edit this part

	err := syscall.Chroot("/host")
	if err != nil {
		log.Println("Failed to create chroot on /host")
		log.Println(err)
		return err
	}

	cmd := exec.Command("/bin/bash", "-c", "systemctl restart "+component)

	err = cmd.Run()

	time.Sleep(10 * time.Second)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func Restart() error {

	err := restartService("etcd")
	if err != nil {
		fmt.Printf("Error restarting etcd: %v\n", err)
	}

	err = restartService("kubelet")
	if err != nil {
		fmt.Printf("Error restarting kubelet: %v\n", err)
	}

	return nil
}
