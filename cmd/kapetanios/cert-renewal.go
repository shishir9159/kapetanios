package main

import (
	"context"
	"fmt"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"log"
	"time"
)

// Step 1. import the pod and create it

func Cert(namespace string) {

	nodeRole := "certs"
	matchLabels := map[string]string{"assigned-node-role.kubernetes.io": "certs"}

	labelSelector := metav1.LabelSelector{MatchLabels: matchLabels}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	// refactor
	client, err := orchestration.NewClient()
	if err != nil {
		fmt.Printf("Error creating Kubernetes client: %v\n", err)
	}

	renewalMinionManager := orchestration.NewMinions(client)

	nodes, err := client.Clientset().CoreV1().Nodes().List(context.Background(), listOptions)
	if err != nil {

	}

	if len(nodes.Items) == 0 {
		log.Fatalln(fmt.Errorf("no master nodes found"))
	}

	for index, node := range nodes.Items {

		// namespace should only be included after the consideration for the existing
		// service account, cluster role binding
		descriptor := renewalMinionManager.MinionBlueprint("quay.io/klovercloud/certs-renewal", nodeRole, node.Name)

		// how many pods this logic need to be in the orchestration too
		minion, er := client.Clientset().CoreV1().Pods(namespace).Create(context.Background(), descriptor, metav1.CreateOptions{})
		if er != nil {
			fmt.Printf("Error creating Cert Renewal pod as the %dth minion: %v\n", index, er)
		}

		fmt.Println(minion)
		fmt.Printf("Cert Renewal pod created as the %dth minion: %s\n", index, minion.Name)

		time.Sleep(5 * time.Second)

		er = RestartByLabel(client, map[string]string{"tier": "control-plane"}, node.Name)
		if er != nil {
			fmt.Println("pod Restart failed")
			break
		}
	}

}
