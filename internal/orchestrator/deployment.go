package orchestrator

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Agent struct {
	client *k8s.Client
}

func NewAgent(client *k8s.Client) *Agent {
	return &Agent{client: client}
}

func (c *Agent) CreateTempPod(ctx context.Context, nodeRole string) (*corev1.Pod, error) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("temp-pod-%s-", nodeRole),
		},
		Spec: corev1.PodSpec{
			NodeSelector: map[string]string{
				"node-role.kubernetes.io/" + nodeRole: "",
			},
			Containers: []corev1.Container{
				{
					Name:  "temp-container",
					Image: "busybox",
					Command: []string{
						"sleep",
						"3600",
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}

	createdPod, err := c.client.Clientset().CoreV1().Pods("default").Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return createdPod, nil
}
