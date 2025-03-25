package main

import (
	"bytes"
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"google.golang.org/grpc"
	"os"
	"os/exec"
)

func compatibility(c Controller, version string, conn *grpc.ClientConn) (bool, error) {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to create chroot on /host")
	}

	errorString := ""

	k8sVersion := "v" + tillHyphen(version)

	// todo: use the configMap
	// --config is not necessary as it is saved in the cm

	//kubeadm upgrade diff to see the changes
	cmd := exec.Command("/bin/bash", "-c", "kubeadm upgrade diff "+k8sVersion+" --config /etc/kubernetes/kubeadm/kubeadm-config.yaml")

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdoutBuf, &stderrBuf

	err = cmd.Run()
	if err != nil {
		c.log.Error().Err(err).
			Msg("error calculating cluster upgrade diff")

		errorString = "error: " + err.Error() + "\n"
		errorString += "error description: " + stderrBuf.String() + "\n"
	}

	// TODO: OS compatibility

	if err = changedRoot(); err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to exit from the updated root")
	}

	conn.ResetConnectBackoff()
	connection := pb.NewMinorUpgradeClient(conn)

	rpc, err := connection.ClusterCompatibility(c.ctx,
		&pb.UpgradeCompatibility{
			OsCompatibility: true,
			Diff:            string(stdoutBuf.Bytes()),
			Err:             errorString,
		})

	c.log.Debug().
		Str("======outStr======", stdoutBuf.String()).
		Msg("diff")

	if err != nil {
		c.log.Error().Err(err).
			Msg("could not send status update")
		// TODO: retry for communication
		return false, err
	}

	c.log.Info().
		Bool("next step", rpc.GetProceedNextStep()).
		Bool("retry current step", rpc.GetRetryCurrentStep()).
		Bool("terminate application", rpc.GetTerminateApplication()).
		Msg("upgrade diff")

	if rpc.GetTerminateApplication() {
		os.Exit(0)
	} else if rpc.GetRetryCurrentStep() {
		return false, nil
	}

	return rpc.GetProceedNextStep(), nil
}
