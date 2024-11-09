package main

import (
	"bytes"
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"go.uber.org/zap"
	"os/exec"
	"time"
)

func restartService(c Controller, component string) (string, string, error) {

	cmd := exec.Command("/bin/bash", "-c", "systemctl restart "+component)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdoutBuf, &stderrBuf

	err := cmd.Run()

	outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	c.log.Info("outString and errString",
		zap.String("outStr", outStr),
		zap.String("errStr", errStr))

	time.Sleep(4 * time.Second)
	if err != nil {
		c.log.Error("Failed to restart service",
			zap.String("component", component),
			zap.Error(err))
		return string(stdoutBuf.Bytes()), string(stderrBuf.Bytes()), err
	}

	return string(stdoutBuf.Bytes()), string(stderrBuf.Bytes()), nil
}

func Restart(c Controller, connection pb.RenewalClient) (bool, bool, error) {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Fatal("Failed to create chroot on /host",
			zap.Error(err))
	}

	etcdLog, etcdErr, err := restartService(c, "etcd")
	if err != nil {
		c.log.Error("Error restarting etcd: ",
			zap.Error(err))
	}

	kubeletLog, kubeletErr, err := restartService(c, "kubelet")
	if err != nil {
		c.log.Error("Error restarting kubelet: ",
			zap.Error(err))
	}

	if err = changedRoot(); err != nil {
		c.log.Fatal("Failed to exit from the updated root",
			zap.Error(err))
	}

	rpc, err := connection.RestartUpdate(c.ctx,
		&pb.RestartStatus{
			EtcdRestart:    true, //TODO:
			KubeletRestart: true, //TODO:
			EtcdLog:        etcdLog,
			EtcdError:      etcdErr,
			KubeletLog:     kubeletLog,
			KubeletError:   kubeletErr,
			Log:            "",
			Err:            "",
		})

	if err != nil {
		c.log.Error("could not send status update: ",
			zap.Error(err))
	}

	c.log.Info("server response",
		zap.Bool("finalizer", rpc.GetGracefullyShutDown()),
		zap.Bool("override existing user kube config", rpc.GetOverrideUserKubeConfig()),
		zap.Bool("retry", rpc.GetRetryRestartingComponents()))

	if rpc.GetRetryRestartingComponents() {
		return false, false, err
	}

	return rpc.GetGracefullyShutDown(), rpc.GetOverrideUserKubeConfig(), nil
}
