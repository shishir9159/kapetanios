package main

import "go.uber.org/zap"

func Upgrade(log *zap.Logger, version string) (bool, error) {

	// The NodeRestriction admission plugin
	// prevents the kubelet from setting or modifying labels
	// with a node-restriction.kubernetes.io/ prefix.

	// example.com.node-restriction.kubernetes.io/
	// role := "minor-upgrade-restriction"

	// control-plane nodes are to be upgraded first

	return false, nil
}
