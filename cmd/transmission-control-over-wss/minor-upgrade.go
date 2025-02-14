package main

// TODO: diff with the original file before merging
//  all todos and remarks had been purged

import (
	"context"
	"encoding/json"
	"fmt"
	jsoniter "github.com/json-iterator/go"
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
	mu        sync.Mutex
	client    *orchestration.Client
	namespace string
	ctx       context.Context
	log       *zap.Logger
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

	labelSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "kapetanios",
		},
	}

	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	namespace := report.minorUpgradeNamespace

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
	configMapName := "kapetanios"
	// TODO: remove
	//var node corev1.Node
	//for _, no := range nodes.Items {
	//	if no.ObjectMeta.Name == "robi-infra-poc-2" {
	//		node = no
	//	}
	//}

	// TODO: refactor this part to orchestrator
	for index, node := range nodes.Items {

		c.log.Info("nodes to be upgraded",
			zap.String("node to be name:", node.Name))

		// todo: populate with user input or not
		targetedVersion := "1.26.6-1.1"
		descriptor := renewalMinionManager.MinionBlueprint("quay.io/klovercloud/minor-upgrade", roleName, node.Name)

		configMap, er := c.client.Clientset().CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
		if er != nil {
			c.log.Error("error fetching the configMap",
				zap.Error(er))
		}

		var nodeNames []string

		for _, no := range nodes.Items[index:] {
			nodeNames = append(nodeNames, no.Name)
		}

		report.nodesToBeUpgraded = nodeNames

		var json = jsoniter.ConfigFastest

		ConfigFastest

		configMap.Data["REPORT"], err = json.Marshal(&report)

		configMap.Data["TARGETED_K8S_VERSION"] = targetedVersion
		configMap.Data["NODES_TO_BE_UPGRADED"] = strings.Join(nodeNames, ";")

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
		go MinorUpgradeGrpc(c.log, pool, ch)

		// TODO: create a watcher against the minion pod

		minion, err := c.client.Clientset().CoreV1().Pods(namespace).Create(context.Background(), descriptor, metav1.CreateOptions{})
		if err != nil {
			c.log.Error("minor upgrade pod creation failed: ",
				zap.Error(err))

			pool.BroadcastMessage([]byte("minor upgrade pod creation failed" + err.Error()))
			return
		}

		c.log.Info("minor upgrade pod created",
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

		defer watcher.Stop()
		if err != nil {
			pool.BroadcastMessage([]byte("failed to create a watcher for the pod: " + minion.Name))

			c.log.Error("failed to create a watcher for the pod",
				zap.Error(err))
			time.Sleep(180 * time.Second)
			(<-ch).Stop()
		}

		// TODO: a loop for all the nodes
		//  wait for the applicationTerminated to be updated

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
					(<-ch).Stop()
					return
				} else if pod.Status.Phase == corev1.PodFailed {
					c.log.Info("minor upgrade pod has failed!",
						zap.String("pod name", pod.Name),
						zap.String("pod namespace", pod.Namespace),
						zap.String("minion name", minion.Name))
					// todo: handle pod failure
					(<-ch).Stop()
					return
				}
			case watch.Deleted:
				fmt.Println("Pod", minion.Name, "was deleted")
				(<-ch).Stop()
				return
			}
		}
	}
}
