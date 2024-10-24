package main

import (
	"github.com/shishir9159/kapetanios/utils"
	"go.uber.org/zap"
	"os/exec"
	"time"
)

func Upgrade(log *zap.Logger, version string) (bool, error) {

	// The NodeRestriction admission plugin
	// prevents the kubelet from setting or modifying labels
	// with a node-restriction.kubernetes.io/ prefix.
	// But, the plugin is needed to be inserted to kubelet configuration
	// and will bring extra complexity.

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		log.Error("Failed to create chroot on /host",
			zap.Error(err))
		return false, err
	}

	// Copy Recursively

	cmd := exec.Command("/bin/bash", "-c", "apt-mark unhold kubeadm && apt-get update && apt-get install -y kubeadm="+version+" && apt-mark hold kubeadm")
	err = cmd.Run()

	time.Sleep(4 * time.Second)
	if err != nil {
		log.Error("Failed to install kubeadm",
			zap.Error(err))
		return false, err
	}

	cmd = exec.Command("/bin/bash", "-c", "kubeadm version")
	err = cmd.Run()

	time.Sleep(4 * time.Second)
	if err != nil {
		log.Error("Failed to install kubeadm",
			zap.Error(err))
		return false, err
	}

	// TODO: --certificate-renewal input from user, and by default should be true
	// TODO: compare the upgrade plan before and after
	cmd = exec.Command("/bin/bash", "-c", "kubeadm upgrade plan --certificate-renewal=true")
	err = cmd.Run()

	// This command checks that your cluster can be upgraded, and fetches the
	// versions you can upgrade to. It also shows a table with the component config
	// version states.

	//TODO: If kubeadm upgrade plan shows any component configs that
	// require manual upgrade, users must provide a config file with replacement
	// configs to kubeadm upgrade apply via the --config command line flag. Failing
	// to do so will cause kubeadm upgrade apply to exit with an error and not
	// perform an upgrade.

	time.Sleep(4 * time.Second)
	if err != nil {
		log.Error("Failed to upgrade kubeadm",
			zap.Error(err))
		return false, err
	}

	// kubeadm version

	//cmd = exec.Command("/bin/bash", "-c", "kubeadm upgrade apply v1.27.x")
	//err = cmd.Run()

	//ime.Sleep(4 * time.Second)
	//	if err != nil {
	//		log.Error("Failed to upgrade kubeadm",
	//			zap.Error(err))
	//		return false, err
	//	}

	// TODO: Same as the first control plane node but use
	//  sudo kubeadm upgrade node
	//  (get this info from an environment value)

	// NOTE: Usage of the --config flag of kubeadm upgrade with kubeadm configuration
	// API types with the purpose of reconfiguring the cluster is not recommended and
	// can have unexpected results. Follow the steps in Reconfiguring a kubeadm
	// cluster instead: https://v1-27.docs.kubernetes.io/docs/tasks/administer-cluster/kubeadm/kubeadm-reconfigure/

	if err = changedRoot(); err != nil {
		log.Fatal("Failed to exit from the updated root",
			zap.Error(err))
	}

	// TODO:
	//  Check for kubernetes repo if no version is found
	//  disableexclude

	return false, nil
}
