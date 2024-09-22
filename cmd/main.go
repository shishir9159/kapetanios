package main

import (
	"context"
	"fmt"
	"github.com/shishir9159/kapetanios/internal/orchestrator"
	"log"
)

func main() {

	client, err := orchestrator.NewClient()
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	renewalAgentManager := orchestrator.NewAgent(client)

	nodeRole := "worker"
	pod, err := renewalAgentManager.CreateTempPod(context.Background(), nodeRole)
	if err != nil {
		log.Fatalf("Error creating temporary pod: %v", err)
	}

	fmt.Printf("Temporary pod created: %s\n", pod.Name)
}
