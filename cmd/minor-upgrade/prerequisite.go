package main

import (
	"bytes"
	"errors"
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"google.golang.org/grpc"
	"os"
	"os/exec"
)

func getDistro(c Controller) (bool, error) {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to create chroot on /host")
	}

	// TODO:
	//  show the changelog of the version by fetching
	//  https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.27.md#changelog-since-v12715

	// TODO: --certificate-renewal input from user, and by default should be true
	// TODO: compare the upgrade plan before and after
	cmd := exec.Command("/bin/bash", "-c", "kubeadm upgrade plan --certificate-renewal=true")

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdoutBuf, &stderrBuf

	err = cmd.Run()

	// This command checks that your cluster can be upgraded, and fetches the
	// versions you can upgrade to. It also shows a table with the component config
	// version states.

	//TODO: If kubeadm upgrade plan shows any component configs that
	// require manual upgrade, users must provide a config file with replacement
	// configs to kubeadm upgrade apply via the --config command line flag. Failing
	// to do so will cause kubeadm upgrade apply to exit with an error and not
	// perform an upgrade.

	if err != nil {
		c.log.Error().Err(err).Msg("failed to calculate upgrade plan")
		// todo: return false, "", err
	}

	//  W1018 10:00:57.703527  599279 common.go:84] your configuration file uses a deprecated API spec: "kubeadm.k8s.io/v1beta2". Please use 'kubeadm config migrate --old-config old.yaml --new-config new.yaml', which will write the new, similar spec using a newer API version.

	// TODO:
	//  show the logs
	//  till these lines come up
	//  [preflight] Running pre-flight checks.

	//[upgrade/config] Reading configuration from the cluster...
	//[upgrade/config] FYI: You can look at this config file with 'kubectl -n kube-system get cm kubeadm-config -o yaml'
	//[preflight] Running pre-flight checks.
	//[upgrade] Running cluster health checks
	//[upgrade] Fetching available versions to upgrade to
	//[upgrade/versions] Cluster version: v1.26.15
	//[upgrade/versions] kubeadm version: v1.26.5
	//I1028 13:48:44.138137 3832104 version.go:256] remote version is much newer: v1.31.2; falling back to: stable-1.26
	//[upgrade/versions] Target version: v1.26.15
	//[upgrade/versions] Latest version in the v1.26 series: v1.26.15

	if err = changedRoot(); err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to exit from the updated root")
	}

	conn.ResetConnectBackoff()

}

func Prerequisites(c Controller, conn *grpc.ClientConn) (bool, error) {

	// TODO: how to know the current node is etcd with clientSet?
	//  	- etcd cluster from the cm
	//		- how to get the hostname and ip address
	//  check if external etcd running if it's an etcd node

	distro, err := getDistro(c)

	etcdNode := os.Getenv("ETCD_NODE")
	if etcdNode == "false" {
		return false, errors.New("ETCD_NODE environment variable set false")
	} else if etcdNode == "true" {
		return false, errors.New("ETCD_NODE environment variable set to be True")
	}

	connection := pb.NewMinorUpgradeClient(conn)

	rpc, err := connection.ClusterHealthChecking(c.ctx,
		&pb.PrerequisitesMinorUpgrade{
			EtcdStatus: true,
			// TODO: refactor
			StorageAvailability: 50,
			Err:                 "",
		})

	if err != nil {
		c.log.Error().Err(err).
			Msg("could not send status update")
		return false, err
	}

	c.log.Info().
		Bool("next step", rpc.GetProceedNextStep()).
		Bool("terminate application", rpc.GetProceedNextStep()).
		Msg("prerequisite step response")

	return rpc.GetProceedNextStep(), nil
}
