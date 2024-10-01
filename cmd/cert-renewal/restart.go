package main

import (
	"context"
	"fmt"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"log"
	"os/exec"
	"syscall"
	"time"
)

func restartByLabel(client *orchestration.Client, matchLabels map[string]string) error {

	// how to add multiple values for one key
	// TO-DO:
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

		time.Sleep(1000000)

		//wait gracefully, for them to restart
	}

	// should I count the retry???

	// watch interface
	for _, pod := range pods.Items {
		po, er := client.Clientset().CoreV1().Pods("kube-system").Get(context.Background(), pod.Name, metav1.GetOptions{})
		if er != nil {
			log.Println(err)
		}

		if po == nil {
			continue
		}

		if po.Status.Phase != corev1.PodRunning {
			// must read and send back error encountered by the restarting pod

			time.Sleep(10 * time.Second)
			continue
		}
		//wait gracefully, for them to restart
	}

	//// stop signal for the informer
	//    stopper := make(chan struct{})
	//    defer close(stopper)
	//
	//    // create shared informers for resources in all known API group versions with a reSync period and namespace
	//    factory := informers.NewSharedInformerFactoryWithOptions(clientSet, 10*time.Second, informers.WithNamespace("demo"))
	//    podInformer := factory.Core().V1().Pods().Informer()
	//
	//    defer runtime.HandleCrash()
	//
	//    // start informer ->
	//    go factory.Start(stopper)
	//
	//    // start to sync and call list
	//    if !cache.WaitForCacheSync(stopper, podInformer.HasSynced) {
	//        runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
	//        return
	//    }
	//
	//    podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
	//        AddFunc:    onAdd, // register add eventhandler
	//        UpdateFunc: onUpdate,
	//        DeleteFunc: onDelete,
	//    })
	//
	//    // block the main go routine from exiting
	//    <-stopper

	return nil
}

func onAdd(obj interface{}) {
	pod := obj.(*corev1.Pod)
	fmt.Printf("POD CREATED: %s/%s", pod.Namespace, pod.Name)
}

func onUpdate(oldObj interface{}, newObj interface{}) {
	oldPod := oldObj.(*corev1.Pod)
	newPod := newObj.(*corev1.Pod)
	fmt.Printf(
		"POD UPDATED. %s/%s %s",
		oldPod.Namespace, oldPod.Name, newPod.Status.Phase,
	)
}

func onDelete(obj interface{}) {
	pod := obj.(*corev1.Pod)
	fmt.Printf("POD DELETED: %s/%s", pod.Namespace, pod.Name)
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

func Restart() error {

	client, err := orchestration.NewClient()
	if err != nil {
		fmt.Printf("Error creating Kubernetes client: %v\n", err)
		return err
	}

	err = restartByLabel(client, map[string]string{"component": "kube-scheduler"})
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
