package utils

import (
	"github.com/gofiber/fiber/v2/log"
	"go.uber.org/zap"
	"os/exec"
)

func Command(command string, args ...string) (string, string, error) {
	changedRoot, err := ChangeRoot("/host")
	if err != nil {
		log.Error("Failed to create chroot on /host",
			zap.Error(err))
		return "", "", err
	}

	cmd := exec.Command("/bin/bash", "-c", "apt-mark unhold ")
	cmd.Run()

	if err = changedRoot(); err != nil {
		log.Fatal("Failed to exit from the updated root",
			zap.Error(err))
	}

	return "", "", nil
}
