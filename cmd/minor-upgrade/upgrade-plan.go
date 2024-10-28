package main

import "go.uber.org/zap"

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

func UpgradePlan(log *zap.Logger, version string) (string, error) {

	// TODO:
	//  show the changelog of the version by fetching
	//  https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.27.md#changelog-since-v12715

	return "", nil
}
