package main

// TODO: diff with the original file before merging
//  all todos and remarks had been purged

import (
	"context"
	"github.com/gorilla/websocket"
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
	"sync"
	"time"
)

var (
	certificateRenewal = false
)

type Controller struct {
	mu        sync.Mutex
	client    *orchestration.Client
	namespace string
	ctx       context.Context
	log       *zap.Logger
}

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

func recovery(namespace string) {

}

func MinorUpgradeFirstRun(namespace string, conn *websocket.Conn) {

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

	if kapetaniosPod == nil {
		if err != nil {
			er := conn.WriteMessage(websocket.TextMessage, []byte("kapetanios pod discovery error"+err.Error()))
			if er != nil {
				c.log.Error("error writing to websocket connection about failed pod discovery error",
					zap.Error(er))
			}
		}

		c.log.Error("check cluster health and communication to kubernetes api server",
			zap.Error(err))
		return
	}

	kapetaniosNode := kapetaniosPod.Items[0].Spec.NodeName
	c.log.Info("kapetanios node",
		zap.String("assigned to the node:", kapetaniosNode))

	nodes, err := c.client.Clientset().CoreV1().Nodes().List(context.Background(), metav1.ListOptions{LabelSelector: ""})

	for _, no := range nodes.Items {
		c.log.Info("nodes",
			zap.String("nodes", no.ObjectMeta.Name))
	}

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
			zap.Error(err))
	}

	roleName := "minor-upgrade"

	var node corev1.Node
	for _, no := range nodes.Items {
		if no.ObjectMeta.Name == "" {
			node = no
		}
	}

	er := conn.WriteMessage(websocket.TextMessage, []byte("kapetanios pod discovery error"+err.Error()))
	if er != nil {
		c.log.Error("error writing to websocket connection about failed pod discovery error",
			zap.Error(er))
	}

	c.log.Info("condition",
		zap.String("node.ObjectMeta.Name", node.ObjectMeta.Name),
		zap.Bool("node.ObjectMeta.Name == kapetaniosNode", node.ObjectMeta.Name == kapetaniosNode))

	configMapName := "kapetanios"

	configMap, er := c.client.Clientset().CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if er != nil {
		c.log.Error("error fetching the configMap",
			zap.Error(er))
	}

	var nodeNames []string

	for _, no := range nodes.Items[index:] {
		nodeNames = append(nodeNames, no.Name)
	}

	//configMap.Data["TARGETED_K8S_VERSION"] = targetedVersion
	configMap.Data["NODES_TO_BE_UPGRADED"] = strings.Join(nodeNames, ";")

	_, er = c.client.Clientset().CoreV1().ConfigMaps(namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
	if er != nil {
		c.log.Error("error updating configMap",
			zap.Error(er))
	}

	c.log.Info("nodes to be upgraded",
		zap.String("node to be", strings.Join(nodeNames, ";")))

	descriptor := renewalMinionManager.MinionBlueprint("quay.io/klovercloud/minor-upgrade", roleName, node.Name)

	descriptor.Spec.Tolerations = []corev1.Toleration{
		{
			Key:      "minor-upgrade-running",
			Operator: "Equal",
			Value:    "processing",
			Effect:   corev1.TaintEffectNoSchedule,
		},
	}

	descriptor.Spec.HostNetwork = true

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

	c.log.Info("cordoning and draining node",
		zap.String("node name", node.Name))

	err = drainAndCordonNode(c, &node)
	if err != nil {
		c.log.Error("failed to drain node",
			zap.String("node name:", node.Name),
			zap.Error(err))
	}

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

	ch := make(chan *grpc.Server, 1)
	go MinorUpgradeGrpc(c.log, conn, ch)

	minion, er := c.client.Clientset().CoreV1().Pods(namespace).Create(context.Background(), descriptor, metav1.CreateOptions{})
	if er != nil {
		c.log.Error("minor upgrade pod creation failed: ",
			zap.Int("index", index),
			zap.Error(er))

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

	(<-ch).Stop()

	removeTaint(&node)
}

func LastDance(c Controller, nodes string, namespace string) {

	renewalMinionManager := orchestration.NewMinions(c.client)

	roleName := "minor-upgrade"

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
		//TODO:
		//  go MinorUpgradeGrpc(c.log, ch)

		minion, er := c.client.Clientset().CoreV1().Pods(namespace).Create(context.Background(), descriptor, metav1.CreateOptions{})
		if er != nil {
			c.log.Error("minor upgrade pod creation failed: ",
				zap.Int("index", index),
				zap.Error(er))

			return
		}

		c.log.Info("minor upgrade pod created",
			zap.Int("index", index),
			zap.String("pod name", minion.Name))

		time.Sleep(50 * time.Second)

		(<-ch).Stop()

		removeTaint(no)
	}
}
