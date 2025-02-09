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
	"strconv"
	"sync"
	"time"
)

var (
	certificateRenewal    = false
	applicationTerminated = false
)

type Controller struct {
	mu        sync.Mutex
	client    *orchestration.Client
	namespace string
	ctx       context.Context
	log       *zap.Logger
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

	if len(nodes.Items) == 0 {
		err = conn.WriteMessage(websocket.TextMessage, []byte("no node found:"+err.Error()))
		if err != nil {
			c.log.Error("no nodes found",
				zap.Error(err))
		}

		return
	}

	roleName := "minor-upgrade"

	var node corev1.Node
	for _, no := range nodes.Items {
		if no.ObjectMeta.Name == "robi-infra-poc-2" {
			node = no
		}
	}

	c.log.Info("condition",
		zap.String("node.ObjectMeta.Name", node.ObjectMeta.Name),
		zap.Bool("node.ObjectMeta.Name == kapetaniosNode", node.ObjectMeta.Name == kapetaniosNode))

	c.log.Info("nodes to be upgraded",
		zap.String("node to be name:", node.Name))

	//targetedVersion := "1.26.6-1.1"
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

	firstNodeToUpgradeEnv := corev1.EnvVar{
		Name:  "FIRST_NODE_TO_BE_UPGRADED",
		Value: "true",
	}

	env = append(env, firstNodeToUpgradeEnv)

	descriptor.Spec.Containers[0].Env = env
	descriptor.Spec.DNSPolicy = corev1.DNSClusterFirstWithHostNet

	ch := make(chan *grpc.Server, 1)
	go MinorUpgradeGrpc(c.log, conn, ch)

	minion, err := c.client.Clientset().CoreV1().Pods(namespace).Create(context.Background(), descriptor, metav1.CreateOptions{})
	if err != nil {
		c.log.Error("minor upgrade pod creation failed: ",
			zap.Error(err))

		er := conn.WriteMessage(websocket.TextMessage, []byte("minor upgrade pod creation failed: "+error.Error(err)))
		if er != nil {
			c.log.Error("failed to write minor upgrade pod creation error in websocket",
				zap.Error(er))
		}

		return
	}

	c.log.Info("minor upgrade pod created",
		zap.String("pod name", minion.Name))

	err = conn.WriteMessage(websocket.TextMessage, []byte("minor upgrade pod created: "+minion.Name))
	if err != nil {
		c.log.Error("failed to write minor upgrade pod creation error in websocket",
			zap.Error(err))
	}

	labelSelector = metav1.LabelSelector{MatchLabels: map[string]string{"app": "minor-upgrade"}}

	listOptions = metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	// TODO: a loop for all the nodes
	//  wait for the applicationTerminated to be updated

	time.Sleep(50 * time.Second)

	(<-ch).Stop()
}
