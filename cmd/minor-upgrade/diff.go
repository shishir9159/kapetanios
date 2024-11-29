package main

import (
	"bytes"
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"google.golang.org/grpc"
	"os/exec"
)

func compatibility(c Controller, version string, conn *grpc.ClientConn) (bool, string, error) {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to create chroot on /host")
	}

	// upgrade plan to list available upgrade options
	// --config is not necessary as it is saved in the cm

	//kubeadm upgrade diff to see the changes
	cmd := exec.Command("/bin/bash", "-c", "kubeadm upgrade diff "+version+" --config /etc/kubernetes/kubeadm/kubeadm-config.yaml")

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdoutBuf, &stderrBuf

	// TODO: CONTAINER-RUNTIME compatibility

	err = cmd.Run()

	if err != nil {
		c.log.Error().Err(err).
			Msg("error calculating cluster upgrade diff")
		// TODO: return the error to the server : return false, "", err
	}

	// TODO: OS compatibility
	diff, _ := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())

	if err = changedRoot(); err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to exit from the updated root")
	}

	conn.ResetConnectBackoff()
	connection := pb.NewMinorUpgradeClient(conn)

	rpc, err := connection.ClusterCompatibility(c.ctx,
		&pb.UpgradeCompatibility{
			OsCompatibility: true,
			Diff:            diff,
			Err:             "", // TODO: check for nil pointer and return
		})

	if err != nil {
		c.log.Error().Err(err).
			Msg("could not send status update")
		// TODO: retry for communication
		return false, "", err
	}

	c.log.Info().
		Bool("next step", rpc.GetProceedNextStep()).
		Bool("terminate application", rpc.GetTerminateApplication()).
		Msg("upgrade diff")

	return rpc.GetProceedNextStep(), diff, nil
}
