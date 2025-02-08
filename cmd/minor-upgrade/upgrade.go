package main

import (
	"bytes"
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"google.golang.org/grpc"
	"os"
	"os/exec"
)

func clusterUpgrade(c Controller, version string, conn *grpc.ClientConn) (bool, error) {

	firstNode := os.Getenv("FIRST_NODE_TO_BE_UPGRADED")
	certRenewal := os.Getenv("CERTIFICATE_RENEWAL")

	//	TODO: kubeadm upgrade node (name) [version] --dry-run

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to create chroot on /host")
	}

	// TODO: get the version number from the upgrade plan
	k8sVersion := "v" + version[:6]

	// TODO: certificate-renewal boolean
	cmd := exec.Command("/bin/bash", "-c", "kubeadm upgrade node --certificate-renewal="+certRenewal+" -y")

	if firstNode == "true" {
		cmd = exec.Command("/bin/bash", "-c", "kubeadm upgrade apply "+k8sVersion+" --certificate-renewal="+certRenewal+" -y")
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdoutBuf, &stderrBuf

	err = cmd.Run()
	//[upgrade/config] Making sure the configuration is correct:
	//[upgrade/config] Reading configuration from the cluster...
	//[upgrade/config] FYI: You can look at this config file with 'kubectl -n kube-system get cm kubeadm-config -o yaml'
	//[preflight] Running pre-flight checks.
	//[upgrade] Running cluster health checks
	//[upgrade/version] You have chosen to change the cluster version to "v1.26.5"
	//[upgrade/versions] Cluster version: v1.26.15
	//[upgrade/versions] kubeadm version: v1.26.5
	//[upgrade/prepull] Pulling images required for setting up a Kubernetes cluster
	//[upgrade/prepull] This might take a minute or two, depending on the speed of your internet connection
	//[upgrade/prepull] You can also perform this action in beforehand using 'kubeadm config images pull'
	//[upgrade/apply] Upgrading your Static Pod-hosted control plane to version "v1.26.5" (timeout: 5m0s)...
	//[upgrade/staticpods] Writing new Static Pod manifests to "/etc/kubernetes/tmp/kubeadm-upgraded-manifests2320749046"
	//[upgrade/staticpods] Preparing for "kube-apiserver" upgrade
	//[upgrade/staticpods] Moved new manifest to "/etc/kubernetes/manifests/kube-apiserver.yaml" and backed up old manifest to "/etc/kubernetes/tmp/kubeadm-backup-manifests-2024-10-28-13-56-36/kube-apiserver.yaml"
	//[upgrade/staticpods] Waiting for the kubelet to restart the component
	//[upgrade/staticpods] This might take a minute or longer depending on the component/version gap (timeout 5m0s)
	//[apiclient] Found 1 Pods for label selector component=kube-apiserver
	//[upgrade/staticpods] Component "kube-apiserver" upgraded successfully!
	//[upgrade/staticpods] Preparing for "kube-controller-manager" upgrade
	//[upgrade/staticpods] Moved new manifest to "/etc/kubernetes/manifests/kube-controller-manager.yaml" and backed up old manifest to "/etc/kubernetes/tmp/kubeadm-backup-manifests-2024-10-28-13-56-36/kube-controller-manager.yaml"
	//[upgrade/staticpods] Waiting for the kubelet to restart the component
	//[upgrade/staticpods] This might take a minute or longer depending on the component/version gap (timeout 5m0s)
	//[apiclient] Found 1 Pods for label selector component=kube-controller-manager
	//[upgrade/staticpods] Component "kube-controller-manager" upgraded successfully!
	//[upgrade/staticpods] Preparing for "kube-scheduler" upgrade
	//[upgrade/staticpods] Moved new manifest to "/etc/kubernetes/manifests/kube-scheduler.yaml" and backed up old manifest to "/etc/kubernetes/tmp/kubeadm-backup-manifests-2024-10-28-13-56-36/kube-scheduler.yaml"
	//[upgrade/staticpods] Waiting for the kubelet to restart the component
	//[upgrade/staticpods] This might take a minute or longer depending on the component/version gap (timeout 5m0s)
	//[apiclient] Found 1 Pods for label selector component=kube-scheduler
	//[upgrade/staticpods] Component "kube-scheduler" upgraded successfully!
	//[upload-config] Storing the configuration used in ConfigMap "kubeadm-config" in the "kube-system" Namespace
	//[kubelet] Creating a ConfigMap "kubelet-config" in namespace kube-system with the configuration for the kubelets in the cluster
	//[kubelet-start] Writing kubelet configuration to file "/var/lib/kubelet/config.yaml"
	//[bootstrap-token] Configured RBAC rules to allow Node Bootstrap tokens to get nodes
	//[bootstrap-token] Configured RBAC rules to allow Node Bootstrap tokens to post CSRs in order for nodes to get long term certificate credentials
	//[bootstrap-token] Configured RBAC rules to allow the csrapprover controller automatically approve CSRs from a Node Bootstrap Token
	//[bootstrap-token] Configured RBAC rules to allow certificate rotation for all node client certificates in the cluster
	//[addons] Applied essential addon: CoreDNS
	//[addons] Applied essential addon: kube-proxy
	//[upgrade/successful] SUCCESS! Your cluster was upgraded to "v1.26.5". Enjoy!
	//[upgrade/kubelet] Now that your control plane is upgraded, please proceed with upgrading your kubelets if you haven't already done so.

	if err = changedRoot(); err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to exit from the updated root")
	}

	conn.ResetConnectBackoff()
	connection := pb.NewMinorUpgradeClient(conn)

	rpc, err := connection.ClusterUpgrade(c.ctx,
		&pb.UpgradeStatus{
			UpgradeSuccess: true,
			Log:            stdoutBuf.String(),
			Err:            "",
		})

	if err != nil {
		c.log.Error().Err(err).
			Msg("could not send status update")
		// TODO: retry for communication
		return false, err
	}

	c.log.Info().
		Bool("next step", rpc.GetProceedNextStep()).
		Bool("terminate application", rpc.GetTerminateApplication()).
		Msg("upgrade status")

	if rpc.GetTerminateApplication() {
		os.Exit(0)
	}

	return rpc.GetProceedNextStep(), nil
}

func k8sComponentsUpgrade(c Controller, k8sComponents string, version string, conn *grpc.ClientConn) (bool, error) {

	//-----// TODO: kernel version compatibility

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to create chroot on /host")
	}

	var upgradeCommand string

	// TODO: use lookup
	if c.distro == "rhel" {
		// TODO: allow unauthenticated
		//upgradeCommand = "yum downgrade -y " + k8sComponents + "-" + version
		upgradeCommand = "yum install -y " + k8sComponents + "-" + version
	} else if c.distro == "ubuntu" {
		upgradeCommand = "apt-mark unhold " + k8sComponents + " && DEBIAN_FRONTEND=noninteractive apt-get install -y " + k8sComponents + "='" + version + "' --allow-unauthenticated && apt-mark hold " + k8sComponents
	}

	// todo: --allow-unauthenticated make it optional
	cmd := exec.Command("/bin/bash", "-c", upgradeCommand)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdoutBuf, &stderrBuf

	err = cmd.Run()

	if err != nil {
		c.log.Error().Err(err).
			Msg("failed to upgrade k8s component")
		// TODO: check updated kubeadm version
		//  return false, err
	}

	outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	c.log.Info().
		Str("outStr", outStr).
		Str("errStr", errStr).
		Msg("outString and errString")

	if err = changedRoot(); err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to exit from the updated root")
	}

	// TODO:
	//  Check for kubernetes repo if no version is found
	//  disableexclude

	conn.ResetConnectBackoff()

	connection := pb.NewMinorUpgradeClient(conn)

	rpc, err := connection.ClusterComponentUpgrade(c.ctx,
		&pb.ComponentUpgradeStatus{
			ComponentUpgradeSuccess: true,
			Component:               k8sComponents,
			Log:                     outStr,
			Err:                     "",
		})

	if err != nil {
		c.log.Error().Err(err).
			Msg("could not send status update")
		// TODO: retry for communication
		return false, err
	}

	c.log.Info().
		Bool("next step", rpc.GetProceedNextStep()).
		Bool("terminate application", rpc.GetTerminateApplication()).
		Msg("upgrade status")

	if rpc.GetTerminateApplication() {
		os.Exit(0)
	}

	return rpc.GetProceedNextStep(), nil
}

// TODO: parse kubeadm upgrade output
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

func upgradePlan(c Controller, conn *grpc.ClientConn) (bool, string, error) {

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
	connection := pb.NewMinorUpgradeClient(conn)
	// todo: send the error with the logs

	rpc, err := connection.ClusterUpgradePlan(c.ctx,
		&pb.UpgradePlan{
			CurrentClusterVersion: "",
			Log:                   "",
			Err:                   "",
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
		Msg("upgrade status")

	if rpc.GetTerminateApplication() {
		os.Exit(0)
	}

	return rpc.GetProceedNextStep(), "", err
}

// TODO: process the output for kubectl and kubelet
// kubectl was already not on hold.
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
//  kubectl kubelet
//2 upgraded, 0 newly installed, 0 to remove and 1 not upgraded.
//Need to get 30.6 MB of archives.
//After this operation, 12.3 kB of additional disk space will be used.
//Get:1 https://prod-cdn.packages.k8s.io/repositories/isv:/kubernetes:/core:/stable:/v1.26/deb  kubectl 1.26.5-1.1 [10.1 MB]
//Get:2 https://prod-cdn.packages.k8s.io/repositories/isv:/kubernetes:/core:/stable:/v1.26/deb  kubelet 1.26.5-1.1 [20.5 MB]
//Fetched 30.6 MB in 2s (15.3 MB/s)
//(Reading database ... 82511 files and directories currently installed.)
//Preparing to unpack .../kubectl_1.26.5-1.1_amd64.deb ...
//Unpacking kubectl (1.26.5-1.1) over (1.26.4-1.1) ...
//Preparing to unpack .../kubelet_1.26.5-1.1_amd64.deb ...
//Unpacking kubelet (1.26.5-1.1) over (1.26.4-1.1) ...
//Setting up kubectl (1.26.5-1.1) ...
//Setting up kubelet (1.26.5-1.1) ...
//Scanning processes...
//Scanning candidates...
//Scanning linux images...
//
//Restarting services...
// systemctl restart kubelet.service
//
//No containers need to be restarted.
//
//No user sessions are running outdated binaries.
//
//No VM guests are running outdated hypervisor (qemu) binaries on this host.

// NOTE: Usage of the --config flag of kubeadm upgrade with kubeadm configuration
// API types with the purpose of reconfiguring the cluster is not recommended and
// can have unexpected results. Follow the steps in Reconfiguring a kubeadm
// cluster instead: https://v1-27.docs.kubernetes.io/docs/tasks/administer-cluster/kubeadm/kubeadm-reconfigure/
