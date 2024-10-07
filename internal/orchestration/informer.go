package orchestration

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"time"
)

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

func SharedInformer(client *kubernetes.Clientset) error {

	//labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"component": "kube-scheduler"}}
	//listOptions := metav1.ListOptions{
	//	LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	//}

	// stop signal for the informer
	stopper := make(chan struct{})
	defer close(stopper)

	// ERROR:
	//  Sadly Shared Informer is not an ideal option

	// create shared informers for resources in all known API group versions with a reSync period, namespace with a specific label

	factory := informers.NewSharedInformerFactoryWithOptions(client, time.Second, informers.WithNamespace("default"))
	//factory := informers.NewSharedInformerFactoryWithOptions(client, time.Second, informers.WithNamespace("default"), listOptions)

	podInformer := factory.Core().V1().Pods().Informer()

	defer runtime.HandleCrash()

	// start informer ->
	go factory.Start(stopper)

	// start to sync and call list
	if !cache.WaitForCacheSync(stopper, podInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		// ToDo: clueless how to handle the error properly
		//return
	}

	handler, err := podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAdd, // register add eventHandler
		UpdateFunc: onUpdate,
		DeleteFunc: onDelete,
	})

	if err != nil {
		if handler != nil {
			if handler.HasSynced() {

			}
		}

		return err
	}

	// block the main go routine from exiting
	<-stopper

	return nil
}

func Informer(client *kubernetes.Clientset, pod *corev1.Pod) error {

	//"component": "kube-scheduler"

	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"tier": "control-plane"}}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		Watch:         true,
	}

	watcher, err := client.CoreV1().Pods("default").Get(context.Background(), pod.Name, metav1.GetOptions{}).Watch(context.Background(), listOptions)
	watcher1, err := client.CoreV1().Pods("default").Watch(context.Background(), listOptions)

	if err != nil {

	}

	if watcher != nil {

	}

	defer watcher.Stop()

	events := watcher.ResultChan()
	for event := range events {
		pod, running := event.Object.(*corev1.Pod)
	}

	return nil
}
