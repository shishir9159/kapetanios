package orchestration

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Agent struct {
	client *Client
}

func NewAgent(client *Client) *Agent {
	return &Agent{client: client}
}

func (c *Agent) CreateTempPod(ctx context.Context, nodeRole string) (*corev1.Pod, error) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("temp-pod-%s-", nodeRole),
			Labels: map[string]string{
				"app": "etcd",
			},
		},
		Spec: corev1.PodSpec{
			Affinity: &corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{
										Key:      "assigned-node-role.kubernetes.io",
										Operator: corev1.NodeSelectorOpIn,
										Values:   []string{nodeRole},
									},
								},
							},
						},
					},
				},
				PodAntiAffinity: &corev1.PodAntiAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
						{
							LabelSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      "assigned-node-role.kubernetes.io",
										Operator: metav1.LabelSelectorOpIn,
										Values:   []string{nodeRole},
									},
								},
							},
							TopologyKey: "kubernetes.io/hostname",
						},
					},
				},
			},
			//NodeSelector: map[string]string{
			//	"assigned-node-role.kubernetes.io": nodeRole,
			//},
			HostPID:     true,
			HostNetwork: true,
			Containers: []corev1.Container{
				{
					Name: "temp-container",
					Command: []string{
						"/bin/bash",
						"-c",
						"chroot /host systemctl status etcd",
					},
					//Image: "quay.io/klovercloud/systemctl-permit:v0.4",
					Image: "ubuntu",
					SecurityContext: &corev1.SecurityContext{
						Privileged: &[]bool{true}[0],
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "host",
							MountPath: "/host",
						},
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
			Volumes: []corev1.Volume{
				{
					Name: "host",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/",
						},
					},
				},
			},
		},
	}

	createdPod, err := c.client.Clientset().CoreV1().Pods("default").Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return createdPod, nil
}
