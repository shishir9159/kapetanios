package main

import (
	"go.uber.org/zap"
	"os/exec"
)

func Diff(log *zap.Logger, s string) (string, error) {

	// upgrade plan to list available upgrade options
	// --config is not necessary as it is saved in the cm

	//kubeadm upgrade diff to see the changes
	cmd := exec.Command("kubeadm", "upgrade diff")

	// TODO:
	//  try combinedOutput and revert back later
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Info("cmd.Run() failed with",
			zap.Error(err))
	}
	log.Info("combined output",
		zap.String("output", string(out)))

	// list all the node names
	// and sort the list from the smallest worker node by resources
	// if it works successfully in the worker nodes, work on the master nodes
	//	--certificate-renewal=false
	//	kubeadm upgrade node (name) [version] --dry-run
	//

	return s, nil
}
