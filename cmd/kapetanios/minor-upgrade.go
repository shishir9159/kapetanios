package main

import (
	"context"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/kubectl/pkg/drain"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

func drainAndCordonNode(c Controller, node *corev1.Node) error {

	drainer := &drain.Helper{
		Ctx:                             c.ctx,
		Client:                          c.client.Clientset(),
		DisableEviction:                 true,
		Force:                           true, // TODO: should it be Force eviction?
		IgnoreAllDaemonSets:             true,
		DeleteEmptyDirData:              true,
		SkipWaitForDeleteTimeoutSeconds: 30,
		Timeout:                         3 * time.Minute,
		GracePeriodSeconds:              10,
		Out:                             os.Stdout,
		ErrOut:                          os.Stderr,
	}

	err := drain.RunCordonOrUncordon(drainer, node, true)
	if err != nil {
		c.log.Error("error cordoning node",
			zap.String("node", node.Name),
			zap.Error(err))

	}

	err = drain.RunNodeDrain(drainer, node.Name)
	if err != nil {
		c.log.Error("error draining node",
			zap.String("node", node.Name),
			zap.Error(err))
	}

	return nil
}

func removeTaint(node *corev1.Node) {

	taints := node.Spec.Taints

	if len(taints) == 0 {
		return
	}

	taintToRemove := corev1.Taint{
		Key:    "minor-upgrade-running",
		Value:  "processing",
		Effect: corev1.TaintEffectNoSchedule,
	}

	var newTaints []corev1.Taint

	for _, taint := range taints {
		if taint.MatchTaint(&taintToRemove) {
			continue
		}
		newTaints = append(newTaints, taint)
	}

	node.Spec.Taints = newTaints
}

func addTaint(node *corev1.Node) {

	taints := node.Spec.Taints

	// TODO: declare as a struct maybe?
	taintToAdd := corev1.Taint{
		Key:    "minor-upgrade-running",
		Value:  "processing",
		Effect: corev1.TaintEffectNoSchedule,
	}

	newTaints := []corev1.Taint{taintToAdd}

	if len(taints) != 0 {
		for _, taint := range taints {
			if taint.MatchTaint(&taintToAdd) {
				return
			}

			newTaints = append(newTaints, taint)
		}

		return
	}

	node.Spec.Taints = newTaints
}

// be careful about the different  across
// the nodes version

// TODO: https://v1-27.docs.kubernetes.io/docs/tasks/administer-cluster/kubeadm/kubeadm-upgrade/#recovering-from-a-failure-state
func recovery(namespace string) {

}

// TODO: for testing purposes, try the current master node 1

func MinorUpgradeFirstRun(namespace string) {

	logger, err := zap.NewProduction()
	defer func(logger *zap.Logger) {
		er := logger.Sync()
		if er != nil {
			logger.Fatal("error syncing logger before application terminates", zap.Error(err))
		}
	}(logger)

	// TODO:
	//  refactor
	client, err := orchestration.NewClient()

	// TODO: add namespace in the controller itself
	c := Controller{
		client: client,
		ctx:    context.Background(),
		log:    logger,
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if err != nil {
		c.log.Error("error creating kubernetes client",
			zap.Error(err))
	}

	renewalMinionManager := orchestration.NewMinions(c.client)

	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"app": "kapetanios"}}

	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	kapetaniosPod, err := c.client.Clientset().CoreV1().Pods(namespace).List(c.ctx, listOptions)

	// assuming there's only one instance of pod
	kapetaniosNode := kapetaniosPod.Items[0].Spec.NodeName
	c.log.Info("kapetanios node",
		zap.String("kapetanios node", kapetaniosNode))

	nodes, err := c.client.Clientset().CoreV1().Nodes().List(context.Background(), metav1.ListOptions{LabelSelector: ""})

	for _, no := range nodes.Items {
		c.log.Info("nodes",
			zap.String("nodes", no.ObjectMeta.Name))
	}

	// TODO: wouldn't work on one master node where lighthouse is scheduled
	// TODO: possible error, lighthouse can be on a master node, that would be mistakenly upgraded at the last

	//	// and sort the list from the smallest worker node by resources

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

	for _, no := range nodes.Items {
		c.log.Info("nodes",
			zap.String("nodes", no.ObjectMeta.Name))
	}

	if err != nil {
		c.log.Error("error listing nodes",
			zap.Error(err))
	}

	if len(nodes.Items) == 0 {
		c.log.Error("no nodes found",
			//	return err or call grpc
			zap.Error(err))

	}

	roleName := "minor-upgrade"

	// TODO: refactor this part to orchestrator

	for index, node := range nodes.Items {

		c.log.Info("condition",
			zap.String("node.ObjectMeta.Name", node.ObjectMeta.Name),
			zap.Bool("node.ObjectMeta.Name == kapetaniosNode", node.ObjectMeta.Name == kapetaniosNode))

		// for reliability purposes, do it for all the nodes
		//if node.ObjectMeta.Name == kapetaniosNode {

		configMapName := "kapetanios"
		//  todo: refactor this hardcoded part
		targetedVersion := "1.26.6-1.1"

		configMap, er := c.client.Clientset().CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
		if er != nil {
			c.log.Error("error fetching the configMap",
				zap.Error(er))
		}

		var nodeNames []string

		for _, no := range nodes.Items[index:] {
			nodeNames = append(nodeNames, no.Name)
		}

		configMap.Data["TARGETED_K8S_VERSION"] = targetedVersion
		configMap.Data["NODES_TO_BE_UPGRADED"] = strings.Join(nodeNames, ";")

		_, er = c.client.Clientset().CoreV1().ConfigMaps(namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
		if er != nil {
			c.log.Error("error updating configMap",
				zap.Error(er))
		}

		c.log.Info("nodes to be upgraded",
			zap.String("node to be", strings.Join(nodeNames, ";")))
		//}

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
		certificateRenewal := false

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
		// todo: query for the kube-dns ip
		descriptor.Spec.DNSConfig = &corev1.PodDNSConfig{
			Nameservers: []string{"10.96.0.1"},
			Searches:    []string{"svc.cluster.local"},
			//Options:     nil,
		}

		// TODO: If any new Pods tolerate the node.kubernetes.io/unschedulable taint,
		//  then those Pods might be scheduled to the node you have drained.

		c.log.Info("cordoning and draining node",
			zap.String("node name", node.Name))

		err = drainAndCordonNode(c, &node)
		if err != nil {
			c.log.Error("failed to drain node",
				zap.String("node name:", node.Name),
				zap.Error(err))
		}

		// TODO:
		//  display the possible errors if force was not

		// TODO:
		//  if ran successfully
		//  Warning: deleting Pods that declare no controller: default/dnsutils; ignoring DaemonSet-managed Pods: kube-system/cilium-wnn6z, kube-system/kube-proxy-m8txw, metallb-system/speaker-587dh
		//evicting pod ingress-nginx/ingress-nginx-admission-create-5pjqv
		//evicting pod metallb-system/controller-9b6c9f6c9-g2p4z
		//evicting pod klovercloud/mesh-uat-go-two-6f846bfcf-ztnk5
		//evicting pod ingress-nginx/ingress-nginx-controller-5dcc84f655-pvdcc
		//evicting pod klovercloud/mesh-uat-go-one-86776497cf-227g9
		//evicting pod default/dnsutils
		//evicting pod default/backend-59b96df495-ghfn2
		//evicting pod ingress-nginx/ingress-nginx-admission-patch-4qgmn
		//evicting pod kube-system/coredns-787d4945fb-stk4w
		//evicting pod kube-system/cilium-operator-fdf6bc9f4-mcx4h
		//evicting pod default/kapetanios-b5bc457fb-v5vlx
		//evicting pod default/minions-for-etcd-4mg95
		//evicting pod kube-system/coredns-787d4945fb-4zl96
		//pod/ingress-nginx-admission-patch-4qgmn evicted
		//pod/ingress-nginx-admission-create-5pjqv evicted
		//pod/controller-9b6c9f6c9-g2p4z evicted
		//pod/minions-for-etcd-4mg95 evicted
		//pod/mesh-uat-go-one-86776497cf-227g9 evicted
		//I1029 03:31:47.241516 4025541 request.go:690] Waited for 1.103702573s due to client-side throttling, not priority and fairness, request: GET:https://10.0.0.3:6443/api/v1/namespaces/klovercloud/pods/mesh-uat-go-two-6f846bfcf-ztnk5
		//pod/dnsutils evicted
		//pod/backend-59b96df495-ghfn2 evicted
		//pod/kapetanios-b5bc457fb-v5vlx evicted
		//pod/mesh-uat-go-two-6f846bfcf-ztnk5 evicted
		//pod/cilium-operator-fdf6bc9f4-mcx4h evicted
		//pod/coredns-787d4945fb-4zl96 evicted
		//pod/coredns-787d4945fb-stk4w evicted
		//pod/ingress-nginx-controller-5dcc84f655-pvdcc evicted

		c.log.Info("tainting node",
			zap.String("node name", node.Name))

		addTaint(&node)

		err = drain.RunCordonOrUncordon(&drain.Helper{
			Ctx:    c.ctx,
			Client: c.client.Clientset(),
		}, &node, false)

		if err != nil {
			c.log.Error("error uncordoning the node",
				zap.String("node name", node.Name),
				zap.Error(err))
		}

		//TODO:
		//ctxTermination, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		//defer stop()
		//
		//var wg sync.WaitGroup
		//
		//// Start the gRPC server in a separate goroutine
		//wg.Add(1)
		//go func() {
		//	defer wg.Done()
		//	MinorUpgradeGrpc(ctxTermination)
		//}()
		//
		//// Wait for the server goroutine to exit
		//<-ctxTermination.Done()
		//stop()
		//wg.Wait()
		//c.log.Info("gRPC server has been gracefully stopped.")

		ch := make(chan *grpc.Server, 1)
		go MinorUpgradeGrpc(c.log, ch)

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
				zap.Int("index", index),
				zap.Error(er))

			//return er
			return
		}

		c.log.Info("minor upgrade pod created",
			zap.Int("index", index),
			zap.String("pod name", minion.Name))

		labelSelector = metav1.LabelSelector{MatchLabels: map[string]string{"app": "minor-upgrade"}}

		listOptions = metav1.ListOptions{
			LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		}

		time.Sleep(50 * time.Second)

		//er = orchestration.Informer(c.client.Clientset(), c.ctx, c.log, 1, listOptions)
		//if er != nil {
		//	c.log.Error("watcher error from minion restart",
		//		zap.Error(er))
		//}

		(<-ch).Stop()

		// TODO: All containers are restarted after upgrade, because the container spec hash value is changed.
		//   check if previously listed pods are all successfully restarted before untainted

		removeTaint(&node)
	}
}

//  ToDo:
//   update the information in the configMaps
//   specially about k8s version

// even though, only one node would be left

func LastDance(c Controller, nodes string, namespace string) {

	renewalMinionManager := orchestration.NewMinions(c.client)

	roleName := "minor-upgrade"

	// TODO: refactor this part to orchestrator

	nodeList := strings.Split(nodes, ";")

	for index, node := range nodeList {

		descriptor := renewalMinionManager.MinionBlueprint("quay.io/klovercloud/minor-upgrade", roleName, node)

		descriptor.Spec.Tolerations = []corev1.Toleration{
			{
				Key:      "minor-upgrade-running",
				Operator: "Equal",
				Value:    "processing",
				Effect:   corev1.TaintEffectNoSchedule,
			},
		}

		descriptor.Spec.HostNetwork = true

		certificateRenewal := false

		certRenewalEnv := corev1.EnvVar{

			Name:  "CERTIFICATE_RENEWAL",
			Value: strconv.FormatBool(certificateRenewal),
		}

		env := descriptor.Spec.Containers[0].Env

		// todo: if the light house got restarted at the first node upgrade: edge case

		env = append(env, certRenewalEnv)
		descriptor.Spec.Containers[0].Env = env
		descriptor.Spec.DNSPolicy = corev1.DNSClusterFirstWithHostNet

		no, err := c.client.Clientset().CoreV1().Nodes().Get(c.ctx, node, metav1.GetOptions{})
		if err != nil {
			c.log.Error("failed to get node by node name",
				zap.String("node name:", node),
				zap.Error(err))
		}

		c.log.Info("cordoning and draining node",
			zap.String("node name", node))

		err = drainAndCordonNode(c, no)
		if err != nil {
			c.log.Error("failed to drain node",
				zap.String("node name:", node),
				zap.Error(err))
		}

		c.log.Info("tainting node",
			zap.String("node name", node))

		addTaint(no)

		err = drain.RunCordonOrUncordon(&drain.Helper{
			Ctx:    c.ctx,
			Client: c.client.Clientset(),
		}, no, false)

		if err != nil {
			c.log.Error("error uncordoning the node",
				zap.String("node name", node),
				zap.Error(err))
		}

		ch := make(chan *grpc.Server, 1)
		go MinorUpgradeGrpc(c.log, ch)

		minion, er := c.client.Clientset().CoreV1().Pods(namespace).Create(context.Background(), descriptor, metav1.CreateOptions{})
		if er != nil {
			c.log.Error("minor upgrade pod creation failed: ",
				zap.Int("index", index),
				zap.Error(er))

			//return er
			return
		}

		c.log.Info("minor upgrade pod created",
			zap.Int("index", index),
			zap.String("pod name", minion.Name))

		(<-ch).Stop()

		removeTaint(no)
	}
}
