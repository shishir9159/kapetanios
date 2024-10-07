package orchestration

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"log"
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

func Informer(client *kubernetes.Clientset, ctx context.Context, l *zap.Logger, listOptions metav1.ListOptions) error {

	// ToDo:
	//	 time limit with context cancellation

	//"component": "kube-scheduler"
	listOptions = metav1.ListOptions{
		LabelSelector: listOptions.LabelSelector,
		FieldSelector: listOptions.FieldSelector,
		Watch:         true,
	}

	watcher, err := client.CoreV1().Pods("default").Watch(context.Background(), listOptions)

	fmt.Println(watcher)

	if err != nil {
		l.Error("error creating the watcher",
			zap.Error(err))
		return err
	}

	if watcher == nil {
		l.Error("watcher is empty")
		fmt.Println(watcher)
		return err
	}

	events := watcher.ResultChan()

	if events == nil {
		fmt.Println(events)
		return fmt.Errorf("event is nil")
	}

	defer watcher.Stop()

	for event := range events {

		pod, running := event.Object.(*corev1.Pod)
		if !running {
			// TODO:
			//	 evicted or pending status check
			log.Printf("pod %s not running %s\n", pod.Name, pod.Status.Phase)
		}

		event.Object.GetObjectKind()

		switch event.Type {
		case watch.Deleted:
			// ToDo: completed status check
			l.Info("pod "+pod.Name+"is deleted",
				zap.Error(nil))
			log.Printf("pod %s is deleted %s\n", pod.Name, pod.Status.Phase)
			return nil
		case watch.Added:
			l.Info("pod "+pod.Name+"is added",
				zap.Error(nil))
			log.Println(pod.Status.ContainerStatuses)
			log.Printf("pod %s is running %s\n", pod.Name, pod.Status.Phase)
			return nil
		case watch.Error:
			log.Printf("error %s\n", event.Object)
			log.Printf("pod %s has failed %s\n", pod.Name, pod.Status.Phase)
			e, _ := client.CoreV1().Events("default").List(ctx, metav1.ListOptions{FieldSelector: "involvedObject.name=" + pod.Name, TypeMeta: metav1.TypeMeta{Kind: "Pod"}})
			return fmt.Errorf(e.String())
		case watch.Modified:
			log.Printf("modified")

		case watch.Bookmark:
			log.Printf("booksmark")

		}
	}

	return nil
}
