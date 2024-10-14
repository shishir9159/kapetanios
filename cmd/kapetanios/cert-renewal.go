package main

import (
	"context"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"time"
)

type Controller struct {
	client *orchestration.Client
	ctx    context.Context
	log    *zap.Logger
}

// excluded etcd servers to be restarted
// etcd, kubelet, control plane component status check
// TODO:
//  etcd-restart

func RestartByLabel(c Controller, matchLabels map[string]string, nodeName string) error {

	// TODO:
	//  how to add multiple values against one key
	labelSelector := metav1.LabelSelector{MatchLabels: matchLabels}
	fieldSelector := metav1.LabelSelector{MatchLabels: map[string]string{"spec.nodeName": nodeName}}

	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		FieldSelector: labels.Set(fieldSelector.MatchLabels).String(),
	}

	minions, err := c.client.Clientset().CoreV1().Pods("kube-system").List(c.ctx, listOptions)
	if err != nil {

		return err
	}

	for _, minion := range minions.Items {
		er := c.client.Clientset().CoreV1().Pods("kube-system").Delete(c.ctx, minion.Name, metav1.DeleteOptions{})
		if er != nil {
			c.log.Info("failed to delete minion:",
				zap.String("minion name:", minion.Name),
				zap.Error(er))
		}
	}

	go func() {
		// todo: instead of the first minion, count the number of minions in switch case
		er := orchestration.Informer(c.client.Clientset(), c.ctx, c.log, len(minions.Items), listOptions)
		if er != nil {
			c.log.Error("watcher error from minion restart",
				zap.Error(er))
		}
	}()

	return nil
}

// etcd restart

func RestartRemainingComponents(c Controller, namespace string) error {

	roleName := "etcd"
	matchLabels := map[string]string{"assigned-etcdNode-role-etcd.kubernetes.io": roleName}

	labelSelector := metav1.LabelSelector{MatchLabels: matchLabels}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	certsNodeQueryListOptions := metav1.ListOptions{
		LabelSelector: labels.Set(map[string]string{"assigned-etcdNode-role-certs.kubernetes.io": "certs"}).String(),
	}

	renewalMinionManager := orchestration.NewMinions(c.client)

	etcdNodes, err := c.client.Clientset().CoreV1().Nodes().List(context.Background(), listOptions)
	if err != nil {
		c.log.Error("error listing etcdNodes",
			zap.Error(err))
	}

	if len(etcdNodes.Items) == 0 {
		c.log.Error("no etcd etcdNodes found",
			zap.Error(err))
		// TODO:
		//  create new error
		return nil
	}

	certNodes, err := c.client.Clientset().CoreV1().Nodes().List(context.Background(), certsNodeQueryListOptions)
	if err != nil {
		c.log.Error("error listing cert etcdNodes",
			zap.Error(err))
	}

	matchFlag := false
	var nodes []string

	for _, etcdNode := range etcdNodes.Items {
		for _, certNode := range certNodes.Items {
			if etcdNode.Name == certNode.Name {
				matchFlag = true
				break
			}
		}

		if matchFlag {
			matchFlag = false
			continue
		}

		nodes = append(nodes, etcdNode.Name)
	}

	for index, node := range nodes {

		// namespace should only be included after the consideration for the existing
		// service account, cluster role binding
		descriptor := renewalMinionManager.MinionBlueprint("quay.io/klovercloud/etcd-restart", roleName, node)

		// kubectl get event --namespace default --field-selector involvedObject.name=minions
		// how many pods this logic need to be in the orchestration too
		minion, er := c.client.Clientset().CoreV1().Pods(namespace).Create(context.Background(), descriptor, metav1.CreateOptions{})
		if er != nil {
			c.log.Error("etcd restart pod creation failed: ",
				zap.Int("index", index),
				zap.Error(er))

			return er
		}

		fieldSelector := metav1.LabelSelector{
			MatchLabels: map[string]string{
				"spec.nodeName": node,
				"metadata.name": minion.Name,
			},
		}

		listOptions = metav1.ListOptions{
			FieldSelector: labels.Set(fieldSelector.MatchLabels).String(),
			LabelSelector: listOptions.LabelSelector,
		}

		er = orchestration.Informer(c.client.Clientset(), c.ctx, c.log, 1, listOptions)
		if er != nil {
			c.log.Error("watcher error from pod restart",
				zap.Error(er))
			return er
		}
	}

	return nil
}

func Cert(namespace string) {

	logger, err := zap.NewProduction()
	defer func(logger *zap.Logger) {
		er := logger.Sync()
		if er != nil {
			logger.Fatal("error syncing logger before application terminates", zap.Error(err))
		}
	}(logger)

	// refactor
	client, err := orchestration.NewClient()

	c := Controller{
		client: client,
		ctx:    context.Background(),
		log:    logger,
	}

	if err != nil {
		c.log.Error("error creating kubernetes client",
			zap.Error(err))
	}

	roleName := "certs"
	matchLabels := map[string]string{"assigned-node-role-certs.kubernetes.io": roleName}

	labelSelector := metav1.LabelSelector{MatchLabels: matchLabels}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	renewalMinionManager := orchestration.NewMinions(client)

	nodes, err := c.client.Clientset().CoreV1().Nodes().List(context.Background(), listOptions)
	if err != nil {
		c.log.Error("error listing nodes",
			zap.Error(err))
	}

	if len(nodes.Items) == 0 {
		c.log.Error("no master nodes found",
			zap.Error(err))
	}

	for index, node := range nodes.Items {

		// namespace should only be included after the consideration for the existing
		// service account, cluster role binding
		descriptor := renewalMinionManager.MinionBlueprint("quay.io/klovercloud/certs-renewal", roleName, node.Name)

		// kubectl get event --namespace default --field-selector involvedObject.name=minions
		// how many pods this logic need to be in the orchestration too
		minion, er := c.client.Clientset().CoreV1().Pods(namespace).Create(context.Background(), descriptor, metav1.CreateOptions{})
		if er != nil {
			c.log.Error("Cert Renewal pod creation failed: ",
				zap.Int("index", index),
				zap.Error(er))

			return
			//return er
		}

		c.log.Info("Cert Renewal pod created",
			zap.Int("index", index),
			zap.String("pod_name", minion.Name))

		// todo: wait for request for restart from the minions
		time.Sleep(5 * time.Second)

		er = RestartByLabel(c, map[string]string{"tier": "control-plane"}, node.Name)
		if er != nil {
			c.log.Error("error restarting pods for certificate renewal",
				zap.Error(er))

			//retry logic
			//return er
			break
		}
	}

	CertGrpc(c.log)

	err = RestartRemainingComponents(c, "default")
	if err != nil {
		c.log.Error("error restarting renewal components", zap.Error(err))
	}
}
