package main

import (
	"context"
	"fmt"
	"github.com/shishir9159/temp-pod-creator/pkg/k8s"
	"github.com/shishir9159/temp-pod-creator/pkg/pod"
	"log"
)

func main() {

	client, err := k8s.NewClient()
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	agent := pod.NewAgent(client)

	nodeRole := "worker"
	pod, err := creator.CreateTempPod(context.Background(), nodeRole)
	if err != nil {
		log.Fatalf("Error creating temporary pod: %v", err)
	}

	fmt.Printf("Temporary pod created: %s\n", pod.Name)
}
