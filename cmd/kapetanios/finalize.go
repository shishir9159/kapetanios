package main

import (
	"context"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"log"
)

// excluded etcd servers to be restarted
// etcd, kubelet, control plane component status check
// TODO:
//  etcd-restart

func RestartByLabel(client *orchestration.Client, matchLabels map[string]string, nodeName string) error {

	// TODO:
	//  how to add multiple values for one key
	labelSelector := metav1.LabelSelector{MatchLabels: matchLabels}
	fieldSelector := metav1.LabelSelector{MatchLabels: map[string]string{"spec.nodeName": nodeName}}

	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		FieldSelector: labels.Set(fieldSelector.MatchLabels).String(),
		//			FieldSelector: "spec.nodeName=" + nodeName,
	}

	pods, err := client.Clientset().CoreV1().Pods("kube-system").List(context.Background(), listOptions)

	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		er := client.Clientset().CoreV1().Pods("kube-system").Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
		if er != nil {
			log.Println(err)
		}
		//
		//time.Sleep(1000000)

		//wait gracefully, for them to restart
	}

	err = orchestration.Informer(client.Clientset(), &labelSelector, &fieldSelector)

	if err != nil {
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

	return nil
}

func Finalize(client *orchestration.Client, nodeName string) {

	//"component": "kube-scheduler"}

}
