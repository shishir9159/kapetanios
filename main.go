package main

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeConfigPath := clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			fmt.Printf("unable to load kubeconfig from %s\n", kubeConfigPath)
		}
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("unable to create clientset from config: %v\n", err)
	}

	nodes, er := clientSet.CoreV1().Nodes().
		List(context.Background(), metav1.ListOptions{LabelSelector: ""})
	if er != nil {
		fmt.Println("error listing nodes")
	}

	for _, node := range nodes.Items {
		if _, controlPlane := node.Annotations["node-role.kubernetes.io/control-plane"]; controlPlane {
			fmt.Println("control plane node", node.Name)
		} else if _, master := node.Annotations["node-role.kubernetes.io/master"]; master {
			fmt.Println("control plane node", node.Name)
		}
	}
}
