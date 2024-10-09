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

	listOptions = metav1.ListOptions{
		TypeMeta:      metav1.TypeMeta{},
		LabelSelector: listOptions.LabelSelector,
		FieldSelector: listOptions.FieldSelector,
		Watch:         true,
	}

	watcher, err := client.CoreV1().Pods("kube-system").Watch(context.Background(), listOptions)

	if err != nil {
		l.Error("error creating the watcher",
			zap.Error(err))
		return err
	}

	defer watcher.Stop()

	// ToDo:
	//  should I count the retry???

	for {
		select {
		case event := <-watcher.ResultChan():

			pod, running := event.Object.(*corev1.Pod)

			l.Info("running")
			if !running {
				l.Error("pod is not running",
					zap.String("pod_name", pod.Name))
				//zap.String("status", pod.Status.Phase))
			}

			switch event.Type {
			case watch.Added:
				l.Info("The pod is added")
				if pod.Status.Phase == corev1.PodRunning {
					//pod.Status.ContainerStatuses
					l.Info("The pod is running")
				}
				fmt.Println(event)
				return nil
			case watch.Modified:
				l.Info("The pod is modified")
				if pod.Status.Phase == corev1.PodRunning {
					l.Info("The pod is running")
				}
				return nil
			case watch.Error:
				e, _ := client.CoreV1().Events("default").List(ctx, metav1.ListOptions{FieldSelector: "involvedObject.name=" + pod.Name, TypeMeta: metav1.TypeMeta{Kind: "Pod"}})
				l.Info("returning event error")
				return fmt.Errorf(e.String())
			}

		//  print watch
		//	{ADDED &Pod{ObjectMeta:{kube-scheduler-shihab-node-1 kube-system
		//	45bdcebb-412a-44b1-9588-ce69f057a21f 25109346 0 2024-10-08 14:03:48 +0000 UTC
		//	2024-10-09 00:31:55 +0000 UTC 0xc0005f9fb0 map[component:kube-scheduler
		//	tier:control-plane]
		//	map[kubernetes.io/config.hash:8517348ed0041f80aef0730b1d4c8053
		//	kubernetes.io/config.mirror:8517348ed0041f80aef0730b1d4c8053
		//	kubernetes.io/config.seen:2024-10-08T08:00:09.229115399Z
		//	kubernetes.io/config.source:file] [{v1 Node shihab-node-1
		//	cd5d4bb0-7a7e-4111-8761-772a1bfa0663 0xc00012a17d <nil>}] [] [{kubelet Update
		//	v1 2024-10-08 14:03:48 +0000 UTC FieldsV1
		//	{"f:metadata":{"f:annotations":{".":{},"f:kubernetes.io/config.hash":{},"f:kubernetes.io/config.mirror":{},"f:kubernetes.io/config.seen":{},"f:kubernetes.io/config.source":{}},"f:labels":{".":{},"f:component":{},"f:tier":{}},"f:ownerReferences":{".":{},"k:{\"uid\":\"cd5d4bb0-7a7e-4111-8761-772a1bfa0663\"}":{}}},"f:spec":{"f:containers":{"k:{\"name\":\"kube-scheduler\"}":{".":{},"f:command":{},"f:image":{},"f:imagePullPolicy":{},"f:livenessProbe":{".":{},"f:failureThreshold":{},"f:httpGet":{".":{},"f:host":{},"f:path":{},"f:port":{},"f:scheme":{}},"f:initialDelaySeconds":{},"f:periodSeconds":{},"f:successThreshold":{},"f:timeoutSeconds":{}},"f:name":{},"f:resources":{".":{},"f:requests":{".":{},"f:cpu":{}}},"f:startupProbe":{".":{},"f:failureThreshold":{},"f:httpGet":{".":{},"f:host":{},"f:path":{},"f:port":{},"f:scheme":{}},"f:initialDelaySeconds":{},"f:periodSeconds":{},"f:successThreshold":{},"f:timeoutSeconds":{}},"f:terminationMessagePath":{},"f:terminationMessagePolicy":{},"f:volumeMounts":{".":{},"k:{\"mountPath\":\"/etc/kubernetes/scheduler.conf\"}":{".":{},"f:mountPath":{},"f:name":{},"f:readOnly":{}}}}},"f:dnsPolicy":{},"f:enableServiceLinks":{},"f:hostNetwork":{},"f:nodeName":{},"f:priorityClassName":{},"f:restartPolicy":{},"f:schedulerName":{},"f:securityContext":{".":{},"f:seccompProfile":{".":{},"f:type":{}}},"f:terminationGracePeriodSeconds":{},"f:tolerations":{},"f:volumes":{".":{},"k:{\"name\":\"kubeconfig\"}":{".":{},"f:hostPath":{".":{},"f:path":{},"f:type":{}},"f:name":{}}}}}
		//	} {kubelet Update v1 2024-10-08 14:03:50 +0000 UTC FieldsV1
		//	{"f:status":{"f:conditions":{".":{},"k:{\"type\":\"ContainersReady\"}":{".":{},"f:lastProbeTime":{},"f:lastTransitionTime":{},"f:status":{},"f:type":{}},"k:{\"type\":\"Initialized\"}":{".":{},"f:lastProbeTime":{},"f:lastTransitionTime":{},"f:status":{},"f:type":{}},"k:{\"type\":\"PodScheduled\"}":{".":{},"f:lastProbeTime":{},"f:lastTransitionTime":{},"f:status":{},"f:type":{}},"k:{\"type\":\"Ready\"}":{".":{},"f:lastProbeTime":{},"f:lastTransitionTime":{},"f:status":{},"f:type":{}}},"f:containerStatuses":{},"f:hostIP":{},"f:phase":{},"f:podIP":{},"f:podIPs":{".":{},"k:{\"ip\":\"5.161.64.103\"}":{".":{},"f:ip":{}}},"f:startTime":{}}}
		//	status}]},Spec:PodSpec{Volumes:[]Volume{Volume{Name:kubeconfig,VolumeSource:VolumeSource{HostPath:&HostPathVolumeSource{Path:/etc/kubernetes/scheduler.conf,Type:*FileOrCreate,},EmptyDir:nil,GCEPersistentDisk:nil,AWSElasticBlockStore:nil,GitRepo:nil,Secret:nil,NFS:nil,ISCSI:nil,Glusterfs:nil,PersistentVolumeClaim:nil,RBD:nil,FlexVolume:nil,Cinder:nil,CephFS:nil,Flocker:nil,DownwardAPI:nil,FC:nil,AzureFile:nil,ConfigMap:nil,VsphereVolume:nil,Quobyte:nil,AzureDisk:nil,PhotonPersistentDisk:nil,PortworxVolume:nil,ScaleIO:nil,Projected:nil,StorageOS:nil,CSI:nil,Ephemeral:nil,},},},Containers:[]Container{Container{Name:kube-scheduler,Image:registry.k8s.io/kube-scheduler:v1.26.15,Command:[kube-scheduler
		//	--authentication-kubeconfig=/etc/kubernetes/scheduler.conf
		//	--authorization-kubeconfig=/etc/kubernetes/scheduler.conf
		//	--bind-address=127.0.0.1 --kubeconfig=/etc/kubernetes/scheduler.conf
		//	--leader-elect=true],Args:[],WorkingDir:,Ports:[]ContainerPort{},Env:[]EnvVar{},Resources:ResourceRequirements{Limits:ResourceList{},Requests:ResourceList{cpu:
		//	{{100 -3} {<nil>} 100m
		//	DecimalSI},},Claims:[]ResourceClaim{},},VolumeMounts:[]VolumeMount{VolumeMount{Name:kubeconfig,ReadOnly:true,MountPath:/etc/kubernetes/scheduler.conf,SubPath:,MountPropagation:nil,SubPathExpr:,},},LivenessProbe:&Probe{ProbeHandler:ProbeHandler{Exec:nil,HTTPGet:&HTTPGetAction{Path:/healthz,Port:{0
		//	10259
		//	},Host:127.0.0.1,Scheme:HTTPS,HTTPHeaders:[]HTTPHeader{},},TCPSocket:nil,GRPC:nil,},InitialDelaySeconds:10,TimeoutSeconds:15,PeriodSeconds:10,SuccessThreshold:1,FailureThreshold:8,TerminationGracePeriodSeconds:nil,},ReadinessProbe:nil,Lifecycle:nil,TerminationMessagePath:/dev/termination-log,ImagePullPolicy:Always,SecurityContext:nil,Stdin:false,StdinOnce:false,TTY:false,EnvFrom:[]EnvFromSource{},TerminationMessagePolicy:File,VolumeDevices:[]VolumeDevice{},StartupProbe:&Probe{ProbeHandler:ProbeHandler{Exec:nil,HTTPGet:&HTTPGetAction{Path:/healthz,Port:{0
		//	10259
		//	},Host:127.0.0.1,Scheme:HTTPS,HTTPHeaders:[]HTTPHeader{},},TCPSocket:nil,GRPC:nil,},InitialDelaySeconds:10,TimeoutSeconds:15,PeriodSeconds:10,SuccessThreshold:1,FailureThreshold:24,TerminationGracePeriodSeconds:nil,},ResizePolicy:[]ContainerResizePolicy{},RestartPolicy:nil,},},RestartPolicy:Always,TerminationGracePeriodSeconds:*30,ActiveDeadlineSeconds:nil,DNSPolicy:ClusterFirst,NodeSelector:map[string]string{},ServiceAccountName:,DeprecatedServiceAccount:,NodeName:shihab-node-1,HostNetwork:true,HostPID:false,HostIPC:false,SecurityContext:&PodSecurityContext{SELinuxOptions:nil,RunAsUser:nil,RunAsNonRoot:nil,SupplementalGroups:[],FSGroup:nil,RunAsGroup:nil,Sysctls:[]Sysctl{},WindowsOptions:nil,FSGroupChangePolicy:nil,SeccompProfile:&SeccompProfile{Type:RuntimeDefault,LocalhostProfile:nil,},},ImagePullSecrets:[]LocalObjectReference{},Hostname:,Subdomain:,Affinity:nil,SchedulerName:default-scheduler,InitContainers:[]Container{},AutomountServiceAccountToken:nil,Tolerations:[]Toleration{Toleration{Key:,Operator:Exists,Value:,Effect:NoExecute,TolerationSeconds:nil,},},HostAliases:[]HostAlias{},PriorityClassName:system-node-critical,Priority:*2000001000,DNSConfig:nil,ShareProcessNamespace:nil,ReadinessGates:[]PodReadinessGate{},RuntimeClassName:nil,EnableServiceLinks:*true,PreemptionPolicy:*PreemptLowerPriority,Overhead:ResourceList{},TopologySpreadConstraints:[]TopologySpreadConstraint{},EphemeralContainers:[]EphemeralContainer{},SetHostnameAsFQDN:nil,OS:nil,HostUsers:nil,SchedulingGates:[]PodSchedulingGate{},ResourceClaims:[]PodResourceClaim{},},Status:PodStatus{Phase:Running,Conditions:[]PodCondition{PodCondition{Type:Initialized,Status:True,LastProbeTime:0001-01-01
		//	00:00:00 +0000 UTC,LastTransitionTime:2024-10-08 08:00:09 +0000
		//	UTC,Reason:,Message:,},PodCondition{Type:Ready,Status:True,LastProbeTime:0001-01-01
		//	00:00:00 +0000 UTC,LastTransitionTime:2024-10-08 08:00:29 +0000
		//	UTC,Reason:,Message:,},PodCondition{Type:ContainersReady,Status:True,LastProbeTime:0001-01-01
		//	00:00:00 +0000 UTC,LastTransitionTime:2024-10-08 08:00:29 +0000
		//	UTC,Reason:,Message:,},PodCondition{Type:PodScheduled,Status:True,LastProbeTime:0001-01-01
		//	00:00:00 +0000 UTC,LastTransitionTime:2024-10-08 08:00:09 +0000
		//	UTC,Reason:,Message:,},},Message:,Reason:,HostIP:5.161.64.103,PodIP:5.161.64.103,StartTime:2024-10-08
		//	08:00:09 +0000
		//	UTC,ContainerStatuses:[]ContainerStatus{ContainerStatus{Name:kube-scheduler,State:ContainerState{Waiting:nil,Running:&ContainerStateRunning{StartedAt:2024-10-08
		//	08:00:10 +0000
		//	UTC,},Terminated:nil,},LastTerminationState:ContainerState{Waiting:nil,Running:nil,Terminated:&ContainerStateTerminated{ExitCode:255,Signal:0,Reason:Unknown,Message:,StartedAt:2024-10-08
		//	07:56:10 +0000 UTC,FinishedAt:2024-10-08 08:00:09 +0000
		//	UTC,ContainerID:containerd://ba1855d830bc6a31958de82d45aca4a9c978c0e23536fc7f9b281863bb48373a,},},Ready:true,RestartCount:26,Image:registry.k8s.io/kube-scheduler:v1.26.15,ImageID:registry.k8s.io/kube-scheduler@sha256:6447dce5ea569c857b161436235292bc30280b3f83fda5df730b23b0812336dc,ContainerID:containerd://72466835538ef8fbc30b54224a6ebb4d8bcd982481edb714b61136c5c04d54cf,Started:*true,AllocatedResources:ResourceList{},Resources:nil,},},QOSClass:Burstable,InitContainerStatuses:[]ContainerStatus{},NominatedNodeName:,PodIPs:[]PodIP{PodIP{IP:5.161.64.103,},},EphemeralContainerStatuses:[]ContainerStatus{},Resize:,ResourceClaimStatuses:[]PodResourceClaimStatus{},HostIPs:[]HostIP{},},}}

		// ToDo:
		//	 time limit with context cancellation
		case <-ctx.Done():
			l.Info("Exit from waitPodRunning for the POD")
			return nil
		}
	}

}
