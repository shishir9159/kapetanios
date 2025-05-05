package main

import (
	"bytes"
	"fmt"
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"google.golang.org/grpc"
	"os"
	"os/exec"
	"strconv"
)

// TODO: remove the previous version packages
//  make sure to check if the previous version exists
//  and current kubernetes version

// TODO: check kubelet status and view the service logs with journalctl -xeu kubelet

type JournalOptions struct {
	Follow     bool
	NumEntries int
	Since      string
	UnitName   string
	FullOutput bool
}

func getJournalLogs(opts JournalOptions) (string, error) {
	var args []string

	args = append(args, "-u", opts.UnitName)

	if opts.Follow {
		args = append(args, "-f")
	}

	if opts.NumEntries > 0 {
		args = append(args, "-n", strconv.Itoa(opts.NumEntries))
	}

	if opts.Since != "" {
		args = append(args, "--since", opts.Since)
	}

	if opts.FullOutput {
		args = append(args, "-x")
	}

	cmd := exec.Command("journalctl", args...)
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get logs for service %s: %w", opts.UnitName, err)
	}

	return out.String(), nil
}

func getServiceLogs(serviceName string) (string, error) {
	cmd := exec.Command("journalctl", "-u", serviceName)
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get logs for service %s: %w", serviceName, err)
	}

	return out.String(), nil
}

func restartComponent(c Controller, component string, conn *grpc.ClientConn) (bool, error) {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to create chroot on /host")
	}

	errorString := ""

	// TODO:
	//  journalctl to get logs

	cmd := exec.Command("/bin/bash", "-c", "systemctl daemon-reload && systemctl restart "+component)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdoutBuf, &stderrBuf

	err = cmd.Run()

	if err != nil {
		c.log.Error().Err(err).
			Msg("failed to restart kubelet")
		errorString = "error: " + err.Error() + "\n"
		errorString += "error description: " + stderrBuf.String() + "\n"
	}

	if err = changedRoot(); err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to exit from the updated root")
	}

	conn.ResetConnectBackoff()
	connection := pb.NewMinorUpgradeClient(conn)

	rpc, err := connection.ClusterComponentRestart(c.ctx,
		&pb.ComponentRestartStatus{
			ComponentRestartSuccess: true,
			Component:               component,
			Log:                     stderrBuf.String(),
			Err:                     errorString,
		})

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
		Msg("backup status")

	if rpc.GetTerminateApplication() {
		os.Exit(0)
	} else if rpc.GetRetryCurrentStep() {
		return false, nil
	}

	return rpc.GetProceedNextStep(), nil
}

//    - `-f` for following logs
//    - `-n` for limiting number of entries
//    - `--since` for time-based filtering
//    - for full output with explanatory messages `-x`

func t() {
	// Basic usage
	logs, err := getServiceLogs("kubelet")
	if err != nil {
		fmt.Println(logs)
		// FATAL
	}

	logs, err = getJournalLogs(JournalOptions{
		UnitName:   "kubelet",
		NumEntries: 100,
		Since:      "1 hour ago",
		FullOutput: true,
	})
	if err != nil {
		// FATAL
	}
}
