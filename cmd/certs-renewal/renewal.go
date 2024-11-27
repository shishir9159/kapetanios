package main

import (
	"bytes"
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"os/exec"
)

//  kubeadm kubeconfig --help
//  export KUBECONFIG=$HOME/.kube/config

//  kubeadm kubeconfig user --org system:nodes --client-name system:node:$(hostname) --config=/etc/kubernetes/kubeadm/kubeadm-config.yaml > /etc/kubernetes/kubelet.conf.new

///usr/local/bin/kubeadm certs renew

func Renew(c Controller, connection pb.RenewalClient) (bool, error) {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to create chroot on /host")
	}

	// whereis kubeadm
	// "/usr/local/bin/kubeadm certs renew scheduler.conf"
	// it is assumed that kubeadm exist otherwise, cert validity wouldn't have work

	// tOdO: is the config location necessary?
	cmd := exec.Command("/usr/bin/kubeadm", "certs", "renew", "all", "--config=/etc/kubernetes/kubeadm/kubeadm-config.yaml")

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdoutBuf, &stderrBuf

	err = cmd.Run()
	if err != nil {
		c.log.Error().Err(err).
			Msg("failed to renew certificates")
	}

	if err = changedRoot(); err != nil {
		c.log.Fatal().Err(err).
			Msg("Failed to exit from the updated root")
	}

	rpc, err := connection.RenewalUpdate(c.ctx,
		&pb.RenewalStatus{
			RenewalSuccess: true,
			RenewalLog:     "",
			RenewalError:   "",
			Log:            "",
			Err:            "",
		})

	if err != nil {
		c.log.Error().Err(err).
			Msg("could not send status update")
	}

	c.log.Info().
		Bool("next step", rpc.GetProceedNextStep()).
		Bool("terminate application", rpc.GetTerminateApplication())

	return rpc.GetProceedNextStep(), err
}
