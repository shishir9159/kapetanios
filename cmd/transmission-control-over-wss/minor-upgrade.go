package main

// TODO: diff with the original file before merging
//  all todos and remarks had been purged

import (
	"context"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"github.com/shishir9159/kapetanios/internal/wss"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	certificateRenewal    = false
	applicationTerminated = false
)

type Controller struct {
	namespace string
	mu        sync.Mutex
	log       *zap.Logger
	ctx       context.Context
	client    *orchestration.Client
}

func recovery(namespace string) {

}

func MinorUpgrade(report *MinorityReport, pool *wss.ConnectionPool) {

	logger := zap.Must(zap.NewProduction())
	defer func(logger *zap.Logger) {
		er := logger.Sync()
		if er != nil {
			logger.Info("error syncing logger before application terminates",
				zap.Error(er))
		}
	}(logger)

	client, err := orchestration.NewClient()

	// TODO: add namespace in the controller itself
	c := Controller{
		client:    client,
		ctx:       context.Background(),
		namespace: "default",
		log:       logger,
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if err != nil {
		c.log.Error("error creating kubernetes client",
			zap.Error(err))
	}

	renewalMinionManager := orchestration.NewMinions(c.client)

	labelSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "kapetanios",
		},
	}

	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	// TODO:
	namespace := "default"
	//namespace := report.MinorUpgradeNamespace

	kapetaniosPod, err := c.client.Clientset().CoreV1().Pods(namespace).List(c.ctx, listOptions)

	if kapetaniosPod == nil {
		if err != nil {
			pool.BroadcastMessage([]byte("kapetanios pod discovery error: " +
				err.Error()))
		}

		c.log.Error("check cluster health and communication to kubernetes api server",
			zap.Error(err))
		return
	}

	kapetaniosNode := kapetaniosPod.Items[0].Spec.NodeName
	c.log.Info("kapetanios node",
		zap.String("assigned to the node:", kapetaniosNode))

	nodes, err := c.client.Clientset().CoreV1().Nodes().List(context.Background(), metav1.ListOptions{LabelSelector: ""})

	if len(nodes.Items) == 0 {
		if err != nil {
			pool.BroadcastMessage([]byte("failed to get node list: " + err.Error()))
		}
		return
	}

	// TODO: debug mode
	//for _, no := range nodes.Items {
	//	c.log.Debug("nodes before sorting",
	//		zap.String("nodes", no.ObjectMeta.Name))
	//}

	// TODO: wouldn't work on one master node where lighthouse is scheduled
	// TODO: possible error, lighthouse can be on a master node, that would be mistakenly upgraded at the last
	//  and sort the list from the smallest worker node by resources
	// TODO: check if node-role.kubernetes.io/control-plane matches with the annotations

	sort.Slice(nodes.Items, func(i, j int) bool {
		if nodes.Items[i].Name == kapetaniosNode {
			return false
		} else if nodes.Items[j].Name == kapetaniosNode {
			return true
		}

		if _, firstNode := nodes.Items[i].Annotations["node-role.kubernetes.io/control-plane"]; firstNode {
			return true
		}
		return false
	})

	// TODO: debug mode
	//for _, no := range nodes.Items {
	//	c.log.Debug("nodes after sorting",
	//		zap.String("nodes", no.ObjectMeta.Name))
	//}

	roleName := "minor-upgrade"
	// todo: configMapName := "kapetanios"

	var nodeNames []string

	for _, no := range nodes.Items {
		c.log.Info("status",
			//zap.String("node config assigned", no.Status.Config.Assigned.String()),
			//zap.String("node config active", no.Status.Config.Active.String()),
			zap.String("node config last known good", no.Status.Config.LastKnownGood.String()),
			zap.String("node condition reason", no.Status.Conditions[0].Reason),
			zap.String("node info os image", no.Status.NodeInfo.OSImage),
			zap.String("node info operating system", no.Status.NodeInfo.OperatingSystem),
			zap.String("node info kernel version", no.Status.NodeInfo.KernelVersion),
			zap.String("node info kubelet version", no.Status.NodeInfo.KubeletVersion),
			zap.String("node info container runtime version", no.Status.NodeInfo.ContainerRuntimeVersion),
		)
		nodeNames = append(nodeNames, no.Name)
	}

	ch := make(chan *grpc.Server, 1)
	go MinorUpgradeGrpc(c.log, pool, ch)

	// TODO: refactor this part to orchestrator
	for index, node := range nodes.Items {

		c.log.Info("nodes to be upgraded",
			zap.String("node to be name:", node.Name))

		// todo: populate with user input or not
		//targetedVersion := "1.26.6-1.1"

		// namespace should only be included after the consideration for the existing
		// service account, cluster role binding
		descriptor := renewalMinionManager.MinionBlueprint("quay.io/klovercloud/minor-upgrade", roleName, node.Name)

		// TODO: instead of pod monitoring for creation, monitor for successful restarts
		//  er = RestartByLabel(c, map[string]string{"tier": "control-plane"}, node.Name)
		//  if er != nil {
		//  	c.log.Error("error restarting pods for certificate renewal",
		//	    	zap.Error(er))
		//	  //retry logic
		//	 // return er
		//	 break
		//  }

		// TODO: drain add node selector or something,
		//   add the same thing on the necessary pods(except for ds)

		//  TODO: after the pod is scheduled
		//   must first drain the node
		//   if failed, must be tainted again to
		//   schedule nodes

		report.NodesToBeUpgraded = strings.Join(nodeNames, ";")
		err = writeConfig(c, *report)
		if err != nil {
			c.log.Error("error writing reporting",
				zap.Error(err))
		}

		descriptor.Spec.Tolerations = []corev1.Toleration{
			{
				Key:      "minor-upgrade-running",
				Operator: "Equal",
				Value:    "processing",
				Effect:   corev1.TaintEffectNoSchedule,
			},
		}

		descriptor.Spec.HostNetwork = true

		// TODO -- take input

		certRenewalEnv := corev1.EnvVar{
			Name:  "CERTIFICATE_RENEWAL",
			Value: strconv.FormatBool(certificateRenewal),
		}

		env := descriptor.Spec.Containers[0].Env
		env = append(env, certRenewalEnv)

		if index == 0 {

			firstNodeToUpgradeEnv := corev1.EnvVar{
				Name:  "FIRST_NODE_TO_BE_UPGRADED",
				Value: "true",
			}

			env = append(env, firstNodeToUpgradeEnv)
		}

		descriptor.Spec.Containers[0].Env = env
		descriptor.Spec.DNSPolicy = corev1.DNSClusterFirstWithHostNet

		//descriptor.Spec.DNSConfig = &corev1.PodDNSConfig{
		//	Nameservers: []string{"10.96.0.10"},
		//	Searches:    []string{"svc.cluster.local"},
		//}

		// TODO: If any new Pods tolerate the node.kubernetes.io/unschedulable taint,
		//  then those Pods might be scheduled to the node you have drained.

		// ----------------- Cordon and Drain ----------------- \\
		//c.log.Info("cordoning and draining node",
		//	zap.String("node name", node.Name))
		//
		//// TODO: should wait for the coredns restart
		//
		//err = drainAndCordonNode(c, &node)
		//if err != nil {
		//	c.log.Error("failed to drain node",
		//		zap.String("node name:", node.Name),
		//		zap.Error(err))
		//}
		//
		//// TODO:
		////  display the possible errors if force was not
		//
		//// TODO:
		////  if ran successfully
		////  Warning: deleting Pods that declare no controller: default/dnsutils; ignoring DaemonSet-managed Pods: kube-system/cilium-wnn6z, kube-system/kube-proxy-m8txw, metallb-system/speaker-587dh
		////evicting pod ingress-nginx/ingress-nginx-admission-create-5pjqv
		////evicting pod metallb-system/controller-9b6c9f6c9-g2p4z
		////evicting pod klovercloud/mesh-uat-go-two-6f846bfcf-ztnk5
		////evicting pod ingress-nginx/ingress-nginx-controller-5dcc84f655-pvdcc
		////evicting pod klovercloud/mesh-uat-go-one-86776497cf-227g9
		////evicting pod default/dnsutils
		////evicting pod default/backend-59b96df495-ghfn2
		////evicting pod ingress-nginx/ingress-nginx-admission-patch-4qgmn
		////evicting pod kube-system/coredns-787d4945fb-stk4w
		////evicting pod kube-system/cilium-operator-fdf6bc9f4-mcx4h
		////evicting pod default/kapetanios-b5bc457fb-v5vlx
		////evicting pod default/minions-for-etcd-4mg95
		////evicting pod kube-system/coredns-787d4945fb-4zl96
		////pod/ingress-nginx-admission-patch-4qgmn evicted
		////pod/ingress-nginx-admission-create-5pjqv evicted
		////pod/controller-9b6c9f6c9-g2p4z evicted
		////pod/minions-for-etcd-4mg95 evicted
		////pod/mesh-uat-go-one-86776497cf-227g9 evicted
		////I1029 03:31:47.241516 4025541 request.go:690] Waited for 1.103702573s due to client-side throttling, not priority and fairness, request: GET:https://10.0.0.3:6443/api/v1/namespaces/klovercloud/pods/mesh-uat-go-two-6f846bfcf-ztnk5
		////pod/dnsutils evicted
		////pod/backend-59b96df495-ghfn2 evicted
		////pod/kapetanios-b5bc457fb-v5vlx evicted
		////pod/mesh-uat-go-two-6f846bfcf-ztnk5 evicted
		////pod/cilium-operator-fdf6bc9f4-mcx4h evicted
		////pod/coredns-787d4945fb-4zl96 evicted
		////pod/coredns-787d4945fb-stk4w evicted
		////pod/ingress-nginx-controller-5dcc84f655-pvdcc evicted
		//
		//c.log.Info("tainting node",
		//	zap.String("node name", node.Name))
		//
		//addTaint(&node)
		//
		//err = drain.RunCordonOrUncordon(&drain.Helper{
		//	Ctx:    c.ctx,
		//	Client: c.client.Clientset(),
		//}, &node, false)
		//
		//if err != nil {
		//	c.log.Error("error uncordoning the node",
		//		zap.String("node name", node.Name),
		//		zap.Error(err))
		//}
		//
		////TODO:
		////ctxTermination, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		////defer stop()
		////
		////var wg sync.WaitGroup
		////
		////// Start the gRPC server in a separate goroutine
		////wg.Add(1)
		////go func() {
		////	defer wg.Done()
		////	MinorUpgradeGrpc(ctxTermination)
		////}()
		////
		////// Wait for the server goroutine to exit
		////<-ctxTermination.Done()
		////stop()
		////wg.Wait()
		////c.log.Info("gRPC server has been gracefully stopped.")
		// ----------------- Cordon and Drain ----------------- \\

		// TODO:
		//  check for pods stuck in the terminating state
		//  if any pods other than the whitelisted ones are still in the node,
		//  force delete

		// TODO:
		//  if the pod doesn't schedule, check for taint
		//  check for all pod related event with informer

		// TODO: monitor the node status with watch

		// TODO: monitor the pod restart after upgrade
		//  All containers are restarted after upgrade, because the container spec hash value is changed
		//		just monitor the NODES before creating minion, no need to restart
		//  RestartByLabel(c, map[string]string{"tier": "control-plane"}, node.Name)

		minion, er := c.client.Clientset().CoreV1().Pods(namespace).Create(context.Background(), descriptor, metav1.CreateOptions{})
		if er != nil {
			c.log.Error("minor upgrade pod creation failed: ",
				zap.Error(er))

			pool.BroadcastMessage([]byte("minor upgrade pod creation failed" + err.Error()))
			return
		}

		c.log.Info("minor upgrade pod created",
			zap.Int("index", index),
			zap.String("pod name", minion.Name))

		pool.BroadcastMessage([]byte("minor upgrade pod created successfully: " + minion.Name))

		labelSelector = metav1.LabelSelector{MatchLabels: map[string]string{"app": "minor-upgrade"}}
		listOptions = metav1.ListOptions{
			LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		}

		watcher, err := c.client.Clientset().CoreV1().Pods(namespace).Watch(context.TODO(), metav1.ListOptions{
			LabelSelector: listOptions.LabelSelector,
			FieldSelector: "metadata.name=" + minion.Name,
		})

		if err != nil {
			pool.BroadcastMessage([]byte("failed to create a watcher for the pod: " + minion.Name))

			c.log.Error("failed to create a watcher for the pod",
				zap.Error(err))
			// handle the error
			time.Sleep(600 * time.Second)
			(<-ch).Stop()
		}

	outerLoop:
		for event := range watcher.ResultChan() {
			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				c.log.Error("watcher returned unexpected type",
					zap.Reflect("event", event),
					zap.Reflect("object", event.Object))
				continue
			}

			switch event.Type {
			case watch.Modified:
				if pod.Status.Phase == corev1.PodSucceeded {
					c.log.Info("minor upgrade pod has completed successfully!",
						zap.String("pod name", pod.Name),
						zap.String("pod namespace", pod.Namespace),
						zap.String("minion name", minion.Name))
					break outerLoop
				} else if pod.Status.Phase == corev1.PodFailed {
					c.log.Info("minor upgrade pod has failed!",
						zap.String("pod name", pod.Name),
						zap.String("pod namespace", pod.Namespace),
						zap.String("minion name", minion.Name))
					// todo: handle pod failure
					break outerLoop
				}
			case watch.Deleted:
				c.log.Info("minor upgrade pod was deleted!",
					zap.String("pod name", pod.Name),
					zap.String("pod namespace", pod.Namespace),
					zap.String("minion name", minion.Name))
				break outerLoop
			}
		}

		// TODO: All containers are restarted after upgrade, because the container spec hash value is changed.
		//  check if previously listed pods are all successfully restarted before untainted

		watcher.Stop()
	}
	(<-ch).Stop()
}
