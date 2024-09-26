package main

import (
	"context"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Step 1. import the pod and create it

func Cert(namespace string) {

	ctx := context.Background()
	client, err := orchestration.NewClient()
	if err != nil {
		//errors.New("Error creating Kubernetes client: %v", err)
	}

	renewalMinionManager := orchestration.NewMinions(client)

	nodeRole := "certs"
	descriptor := renewalMinionManager.MinionBlueprint("ubuntu", nodeRole)

	// how many pods
	// this logic need to be in the orchestration too
	minion, err := client.Clientset().CoreV1().Pods(namespace).Create(ctx, descriptor, metav1.CreateOptions{})
	if err != nil {
		//fmt.Println("Error creating temporary pod: %v", err)
	}

	print(minion)

	//		fmt.Printf("Temporary pod created: %s\n", pod.Name)
}
