package main

import (
	"bytes"
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"os/exec"
)

func restartService(c Controller, component string) (string, string, error) {

	cmd := exec.Command("/bin/bash", "-c", "systemctl restart "+component)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdoutBuf, &stderrBuf

	err := cmd.Run()

	outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	c.log.Info().
		Str("outStr", outStr).
		Str("errStr", errStr).
		Msg("outString and errString")

	if err != nil {
		c.log.Error().Err(err).
			Str("component", component).
			Msg("failed to restart service")
		return string(stdoutBuf.Bytes()), string(stderrBuf.Bytes()), err
	}

	return string(stdoutBuf.Bytes()), string(stderrBuf.Bytes()), nil
}

func Restart(c Controller, connection pb.RenewalClient) (bool, bool, error) {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to create chroot on /host")
	}

	etcdLog, etcdErr, err := restartService(c, "etcd")
	if err != nil {
		c.log.Error().Err(err).
			Msg("failed to restart etcd")
	}

	kubeletLog, kubeletErr, err := restartService(c, "kubelet")
	if err != nil {
		c.log.Error().Err(err).
			Msg("failed to restart kubelet")
	}

	if err = changedRoot(); err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to exit from the updated root")
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
		c.log.Error().Err(err).
			Msg("could not send status update")
	}

	c.log.Info().
		Bool("retry", rpc.GetRetryCurrentStep()).
		Bool("finalizer", rpc.GetResponseReceived()).
		Bool("override existing user kube config", rpc.GetOverrideUserKubeConfig()).
		Msg("server response")

	overrideUserKubeConfig := rpc.GetOverrideUserKubeConfig()

	if rpc.GetRetryCurrentStep() {
		overrideUserKubeConfig = false
	}

	return rpc.GetRetryCurrentStep(), overrideUserKubeConfig, nil
}
