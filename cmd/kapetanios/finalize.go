package main

import (
	"fmt"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"log"
)

// excluded etcd servers to be restarted
// etcd, kubelet, control plane component status check
// TODO:
//  etcd-restart

func RestartByLabel(c Controller, matchLabels map[string]string, nodeName string) error {

	// TODO:
	//  how to add multiple values for one key
	labelSelector := metav1.LabelSelector{MatchLabels: matchLabels}
	fieldSelector := metav1.LabelSelector{MatchLabels: map[string]string{"spec.nodeName": nodeName}}

	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		FieldSelector: labels.Set(fieldSelector.MatchLabels).String(),
		//			FieldSelector: "spec.nodeName=" + nodeName,
	}

	c.log.Info("restart by label")

	pods, err := c.client.CoreV1().Pods("kube-system").List(c.ctx, listOptions)
	if err != nil {

		return err
	}

	for _, pod := range pods.Items {
		er := c.client.CoreV1().Pods("kube-system").Delete(c.ctx, pod.Name, metav1.DeleteOptions{})
		if er != nil {
			log.Println(err)
		}
		//
		//time.Sleep(1000000)

		//wait gracefully, for them to restart
	}

	//listOptions

	err = orchestration.Informer(c.client, c.ctx, c.log, listOptions)

	if err != nil {
		fmt.Println("orchestration informer error:")
		fmt.Println(err)
		return err
	}

	// ToDo:
	//  should I count the retry???

	// watch interface
	//for _, pod := range pods.Items {
	//	po, er := client.Clientset().CoreV1().Pods("kube-system").Get(context.Background(), pod.Name, metav1.GetOptions{})
	//	if er != nil {
	//		log.Println(err)
	//	}
	//
	//	if po == nil {
	//		continue
	//	}
	//
	//	if po.Status.Phase != corev1.PodRunning {
	//		// must read and send back error encountered by the restarting pod
	//
	//		time.Sleep(10 * time.Second)
	//		continue
	//	}
	//}

	defer func(log *zap.Logger) {
		er := log.Sync()
		if er != nil {

		}
	}(c.log)
	return nil
}

func Finalize(client *orchestration.Client, nodeName string) {

	//"component": "kube-scheduler"}

}
