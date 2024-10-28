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

	//var Stdin io.Reader
	//var Stdout io.Writer
	//var Stderr io.Writer

	cmd := exec.Command("/bin/bash", "-c", "apt-mark unhold kubeadm && apt-get update && apt-get install -y kubeadm="+version+" && apt-mark hold kubeadm")
	err = cmd.Run()

	// TODO: possible output format
	//kubeadm was already not on hold.
	//Hit:1 https://mirror.hetzner.com/ubuntu/packages jammy InRelease
	//Hit:2 https://mirror.hetzner.com/ubuntu/packages jammy-updates InRelease
	//Hit:3 https://mirror.hetzner.com/ubuntu/packages jammy-backports InRelease
	//Hit:4 https://mirror.hetzner.com/ubuntu/security jammy-security InRelease
	//Hit:5 https://download.docker.com/linux/ubuntu jammy InRelease
	//Hit:6 https://prod-cdn.packages.k8s.io/repositories/isv:/kubernetes:/core:/stable:/v1.26/deb  InRelease
	//Reading package lists... Done
	//Reading package lists... Done
	//Building dependency tree... Done
	//Reading state information... Done
	//The following packages will be upgraded:
	//  kubeadm
	//1 upgraded, 0 newly installed, 0 to remove and 2 not upgraded.
	//Need to get 9,746 kB of archives.
	//After this operation, 4,096 B of additional disk space will be used.
	//Get:1 https://prod-cdn.packages.k8s.io/repositories/isv:/kubernetes:/core:/stable:/v1.26/deb  kubeadm 1.26.5-1.1 [9,746 kB]
	//Fetched 9,746 kB in 2s (4,044 kB/s)
	//(Reading database ... 82511 files and directories currently installed.)
	//Preparing to unpack .../kubeadm_1.26.5-1.1_amd64.deb ...
	//Unpacking kubeadm (1.26.5-1.1) over (1.26.4-1.1) ...
	//Setting up kubeadm (1.26.5-1.1) ...
	//Scanning processes...
	//Scanning linux images...
	//
	//No services need to be restarted.
	//
	//No containers need to be restarted.
	//
	//No user sessions are running outdated binaries.
	//
	//No VM guests are running outdated hypervisor (qemu) binaries on this host.
	//kubeadm set on hold.

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
