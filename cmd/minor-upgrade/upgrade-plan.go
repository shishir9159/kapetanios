package main

import "go.uber.org/zap"

//  W1018 10:00:57.703527  599279 common.go:84] your configuration file uses a deprecated API spec: "kubeadm.k8s.io/v1beta2". Please use 'kubeadm config migrate --old-config old.yaml --new-config new.yaml', which will write the new, similar spec using a newer API version.

// TODO:
//  show the logs
//  till these lines come up
//  [preflight] Running pre-flight checks.
//[upgrade] Running cluster health checks
//[upgrade] Fetching available versions to upgrade to
//[upgrade/versions] Cluster version: v1.26.15
//[upgrade/versions] kubeadm version: v1.26.4
//[upgrade/versions] Target version: v1.26.15
//[upgrade/versions] Latest version in the v1.26 series: v1.26.15

func UpgradePlan(log *zap.Logger) {

	// TODO:
	//  Check which versions are available to upgrade to and
	//  validate whether your current cluster is upgradeable

	log.Info("")

	return
}
