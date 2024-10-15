package main

import (
	"context"
	"github.com/gofiber/fiber/v2/log"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

// store certificate validity
// check number of nodes
// save certDir to configMap
// checking if the necessary files exist

//etcd:
//  external:
//    caFile: /etc/kubernetes/pki/etcd-ca.pem
//    certFile: /etc/kubernetes/pki/etcd.cert
//    endpoints:
//    - https://5.161.64.103:2379
//    - https://5.161.248.112:2379
//    - https://5.161.67.249:2379
//    keyFile: /etc/kubernetes/pki/etcd.key
//kubernetesVersion

// format Data:map[string]string{ClusterConfiguration

func populatingConfigMap(c Controller) error {

	cm, err := c.client.Clientset().CoreV1().ConfigMaps("kube-system").Get(context.Background(), "kubeadm-config", metav1.GetOptions{})

	if err != nil {
		c.log.Error("error fetching the kubeadm-config from the kube-system namespace", zap.Error(err))
		return err
	}

	// ClusterConfiguration stores the kubeadm-config as a file in the configmap

	configSlice := strings.Split(cm.Data["ClusterConfiguration"], "\n")

	for index, config := range configSlice {
		log.Info(zap.Int("index", index),
			zap.String("config", config))
	}

	return nil
}

func InitialSetup(c Controller) {

	err := populatingConfigMap(c)
	if err != nil {
		c.log.Error("error populating config", zap.Error(err))
	}
}
