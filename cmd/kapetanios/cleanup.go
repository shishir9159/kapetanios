package main

import (
	"context"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func Cleanup(namespace string) error {

	logger, err := zap.NewProduction()
	defer func(logger *zap.Logger) {
		er := logger.Sync()
		if er != nil {
			logger.Info("error syncing logger before application terminates", zap.Error(err))
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
		"app": "certs",
	}

	labelSelector := metav1.LabelSelector{MatchLabels: matchLabels}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	var minions []corev1.Pod

	pods, err := c.client.Clientset().CoreV1().Pods(namespace).List(context.Background(), listOptions)
	if err != nil {
		c.log.Error("error listing minions for certificate renewals",
			zap.Error(err))
		return err
	}

	minions = pods.Items

	if len(minions) == 0 {
		c.log.Error("no completed minions found")
	}

	secondMatchLabels := map[string]string{
		"app": "etcd",
	}

	labelSelector = metav1.LabelSelector{MatchLabels: secondMatchLabels}
	listOptions = metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	pods, err = c.client.Clientset().CoreV1().Pods(namespace).List(context.Background(), listOptions)
	if err != nil {
		c.log.Error("error listing minions for etcd cleanup",
			zap.Error(err))
		return err
	}

	if len(pods.Items) == 0 {
		c.log.Info("no completed minions found")
	}

	minions = append(minions, pods.Items...)

	deletePolicy := metav1.DeletePropagationForeground

	for _, minion := range minions {
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

	return nil
}
