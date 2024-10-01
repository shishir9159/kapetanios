package minor_upgrade

import (
	"fmt"
	"log"
	"os/exec"
)

func main() {
	// upgrade plan to list available upgrade options
	// --config is not necessary as it is saved in the cm

	//kubeadm upgrade diff to see the changes
	cmd := exec.Command("kubeadm", "upgrade diff")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	fmt.Printf("combined out:\n%s\n", string(out))

	// list all the node names
	// and sort the list from the smallest worker node by resources
	// if it works successfully in the worker nodes, work on the master nodes
	//	--certificate-renewal=false
	//	kubeadm upgrade node (name) [version] --dry-run
	//
}
