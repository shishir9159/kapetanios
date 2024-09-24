package main

import (
	"context"
	"fmt"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"os"
)

func main() {

	client, err := orchestration.NewClient()
	if err != nil {
		fmt.Println("Error creating Kubernetes client: %v", err)
	}

	renewalAgentManager := orchestration.NewAgent(client)

	nodeRole := "etcd"
	pod, err := renewalAgentManager.CreateTempPod(context.Background(), nodeRole)
	if err != nil {
		fmt.Println("Error creating temporary pod: %v", err)
	}

	defer func() {
		fmt.Printf("Temporary pod created: %s\n", pod.Name)
	}()

	os.Exit(0)
}
