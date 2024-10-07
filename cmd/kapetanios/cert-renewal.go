package main

import (
	"context"
	"fmt"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"time"
)

type Controller struct {
	client *kubernetes.Clientset
	ctx    context.Context
	log    *zap.Logger
}

// Step 1. import the pod and create it

func Cert(namespace string) {

	logger, err := zap.NewProduction()
	defer func(logger *zap.Logger) {
		er := logger.Sync()
		if er != nil {

		}
	}(logger)

	// refactor
	client, err := orchestration.NewClient()

	c := Controller{
		client: client.Clientset(),
		ctx:    context.Background(),
		log:    logger,
	}

	if err != nil {
		c.log.Error("error creating kubernetes client",
			zap.Error(err))
	}

	nodeRole := "certs"
	matchLabels := map[string]string{"assigned-node-role.kubernetes.io": "certs"}

	labelSelector := metav1.LabelSelector{MatchLabels: matchLabels}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	renewalMinionManager := orchestration.NewMinions(client)

	nodes, err := c.client.CoreV1().Nodes().List(context.Background(), listOptions)
	if err != nil {
		c.log.Error("error listing nodes",
			zap.Error(err))
	}

	if len(nodes.Items) == 0 {
		c.log.Error("no master nodes found",
			zap.Error(err))
	}

	for index, node := range nodes.Items {

		// namespace should only be included after the consideration for the existing
		// service account, cluster role binding
		descriptor := renewalMinionManager.MinionBlueprint("quay.io/klovercloud/certs-renewal", nodeRole, node.Name)

		// how many pods this logic need to be in the orchestration too
		minion, er := c.client.CoreV1().Pods(namespace).Create(context.Background(), descriptor, metav1.CreateOptions{})
		if er != nil {
			fmt.Printf("Error creating Cert Renewal pod as the %dth minion: %v\n", index, er)
		}

		fmt.Printf("Cert Renewal pod created as the %dth minion: %s\n", index, minion.Name)

		time.Sleep(5 * time.Second)

		er = RestartByLabel(c, map[string]string{"tier": "control-plane"}, node.Name)
		if er != nil {
			fmt.Println("pod Restart failed")
			break
		}

	}

}
