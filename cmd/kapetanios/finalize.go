package main

import (
	"context"
	"fmt"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

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

	c.log.Debug("entered restart remaining components")

	roleName := "etcd"
	matchLabels := map[string]string{"assigned-node-role-etcd.kubernetes.io": roleName}

	labelSelector := metav1.LabelSelector{MatchLabels: matchLabels}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	certsNodeQueryListOptions := metav1.ListOptions{
		LabelSelector: labels.Set(map[string]string{"assigned-node-role-certs.kubernetes.io": "certs"}).String(),
	}

	renewalMinionManager := orchestration.NewMinions(c.client)

	c.log.Debug("listing etcd nodes")

	etcdNodes, err := c.client.Clientset().CoreV1().Nodes().List(context.Background(), listOptions)
	if err != nil {
		c.log.Error("error listing etcd nodes",
			zap.Error(err))
	}

	if len(etcdNodes.Items) == 0 {
		c.log.Error("no etcd etcd nodes found",
			zap.Error(err))
		// TODO:
		//  create new error
		return nil
	}

	certNodes, err := c.client.Clientset().CoreV1().Nodes().List(context.Background(), certsNodeQueryListOptions)
	if err != nil {
		c.log.Error("error listing cert nodes",
			zap.Error(err))
	}

	matchFlag := false
	var nodes []string

	fmt.Println(certNodes)
	fmt.Println(etcdNodes)

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

	c.log.Debug("listing all compliment set of nodes",
		zap.String("nodes[0]", nodes[0]))

	for index, node := range nodes {

		// namespace should only be included after the consideration for the existing
		// service account, cluster role binding
		//descriptor := renewalMinionManager.MinionBlueprint("quay.io/klovercloud/etcd-restart", roleName, node)

		// TODO:
		//  Update the Dockerfile
		descriptor := renewalMinionManager.MinionBlueprint("quay.io/klovercloud/systemctl-permit:v0.4", roleName, node)

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
