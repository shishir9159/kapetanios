package main

import (
	"context"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func Cleanup(namespace string) {

	logger, err := zap.NewProduction()
	defer func(logger *zap.Logger) {
		er := logger.Sync()
		if er != nil {
			logger.Fatal("error syncing logger before application terminates", zap.Error(err))
		}
	}(logger)

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

	matchLabels := map[string]string{
		"assigned-node-role-certs.kubernetes.io": "certs",
		"assigned-node-role-etcd.kubernetes.io":  "etcd",
	}

	labelSelector := metav1.LabelSelector{MatchLabels: matchLabels}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	minions, err := c.client.Clientset().CoreV1().Pods(namespace).List(context.Background(), listOptions)
	if err != nil {
		c.log.Error("error listing pods",
			zap.Error(err))
	}

	c.log.Info("pods",
		zap.String("minions.Items", minions.Items[0].Name))

	if len(minions.Items) == 0 {
		c.log.Error("no completed minions found",
			zap.Error(err))
	}

	deletePolicy := metav1.DeletePropagationForeground

	for _, minion := range minions.Items {
		er := c.client.Clientset().CoreV1().Pods(namespace).Delete(c.ctx, minion.Name, metav1.DeleteOptions{
			GracePeriodSeconds: &[]int64{3}[0],
			PropagationPolicy:  &deletePolicy,
		})
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

	//	return err
}
