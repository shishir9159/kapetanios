package main

import (
	"context"
	"fmt"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"log"
	"os/exec"
	"syscall"
	"time"
)

func restartByLabel(client *orchestration.Client, matchLabels map[string]string) error {

	// TODO:
	//  how to add multiple values for one key
	labelSelector := metav1.LabelSelector{MatchLabels: matchLabels}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
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

	err = orchestration.Informer(client.Clientset(), matchLabels)

	if err != nil {

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

func restartService(component string) error {

	err := syscall.Chroot("/host")
	if err != nil {
		log.Println("Failed to create chroot on /host")
		log.Println(err)
		return err
	}

	cmd := exec.Command("/bin/bash", "-c", "systemctl restart "+component)

	err = cmd.Run()

	time.Sleep(10 * time.Second)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func Restart(client *orchestration.Client) error {

	//"component": "kube-scheduler"}

	err := restartByLabel(client, map[string]string{"tier": "control-plane"})
	if err != nil {
		fmt.Printf("Error restarting kube-scheduler: %v\n", err)
	}

	err = restartService("etcd")
	if err != nil {
		fmt.Printf("Error restarting etcd: %v\n", err)
	}

	err = restartService("kubelet")
	if err != nil {
		fmt.Printf("Error restarting kubelet: %v\n", err)
	}

	return nil
}
