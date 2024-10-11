package orchestration

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Minions struct {
	client *Client
}

func NewMinions(client *Client) *Minions {
	return &Minions{client: client}
}

func (c *Minions) MinionBlueprint(image string, nodeRole string, nodeName string) *corev1.Pod {

	blueprint := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("minions-for-%s-", nodeRole),
			//Namespace: namespace,
			Labels: map[string]string{
				"app": nodeRole,
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
			NodeName:    nodeName,
			Containers: []corev1.Container{
				{
					Name:  "certs-renewal",
					Image: image,
					SecurityContext: &corev1.SecurityContext{
						Privileged: &[]bool{true}[0],
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "host",
							MountPath: "/host",
						},
					},
					Env: []corev1.EnvVar{
						{
							Name:  "GRPC_DNS_RESOLVER",
							Value: "native",
						},
						{
							Name:  "GRPC_GO_LOG_SEVERITY_LEVEL",
							Value: "INFO",
						},
						{
							Name:  "GRPC_GO_LOG_VERBOSITY_LEVEL",
							Value: "99",
						},
						{
							Name:  "GRPC_TRACE",
							Value: "all",
						},
						{
							Name:  "GODEBUG",
							Value: "http2debug=2",
						},
					},
				},
			},
			//DNSPolicy: corev1.DNSClusterFirst,
			// Todo:
			//  fix on the grpc client rather than the descriptor
			//DNSConfig: &corev1.PodDNSConfig{
			//	Nameservers: []string{"10.96.0.10"},
			//},
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

	return blueprint
}
