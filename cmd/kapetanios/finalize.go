package main

import (
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// excluded etcd servers to be restarted
// etcd, kubelet, control plane component status check
// TODO:
//  etcd-restart

func RestartByLabel(c Controller, matchLabels map[string]string, nodeName string) error {

	// TODO:
	//  how to add multiple values against one key
	labelSelector := metav1.LabelSelector{MatchLabels: matchLabels}
	fieldSelector := metav1.LabelSelector{MatchLabels: map[string]string{"spec.nodeName": nodeName}}

	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		FieldSelector: labels.Set(fieldSelector.MatchLabels).String(),
	}

	pods, err := c.client.CoreV1().Pods("kube-system").List(c.ctx, listOptions)
	if err != nil {

		return err
	}

	for _, pod := range pods.Items {
		er := c.client.CoreV1().Pods("kube-system").Delete(c.ctx, pod.Name, metav1.DeleteOptions{})
		if er != nil {
			c.log.Info("failed to delete pod:",
				zap.String("pod name:", pod.Name),
				zap.Error(er))
		}
	}

	go func() {
		// todo: instead of the first pod, count the number of pods in switch case
		er := orchestration.Informer(c.client, c.ctx, c.log, len(pods.Items), listOptions)
		if er != nil {
			c.log.Error("watcher error from pod restart",
				zap.Error(er))
		}
	}()

	return nil
}

func Finalize(client *orchestration.Client, nodeName string) {

	//"component": "kube-scheduler"}

}
