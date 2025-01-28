package orchestration

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	clientSet *kubernetes.Clientset
}

func NewClient() (*Client, error) {

	config, err := rest.InClusterConfig()
	if err != nil {
		kubeConfigPath := clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			return nil, fmt.Errorf(
				"unable to load kubeconfig from %s: %v",
				kubeConfigPath, err)
		}
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Client{clientSet: clientSet}, nil
}

func (c *Client) Clientset() *kubernetes.Clientset {
	return c.clientSet
}
