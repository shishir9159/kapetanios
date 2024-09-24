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

//apiVersion: v1
//kind: Pod
//metadata:
//  name: service-control-pod
//spec:
//  nodeName: shihab-node-1
//  containers:
//  - name: privileged-container
//    image: quay.io/klovercloud/systemctl-permit:v0.4
//    securityContext:
//      privileged: true
//    command:
//    - "/bin/bash"
//    - "-c"
//    - "chroot /host systemctl status etcd"
//    - "echo 'command executed'"
//    - "sleep 10m"
//    volumeMounts:
//    - mountPath: /host
//      name: host
//  volumes:
//  - name: host
//    hostPath:
//      path: /

func (c *Agent) CreateTempPod(ctx context.Context, nodeRole string) (*corev1.Pod, error) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("temp-pod-%s-", nodeRole),
		},
		Spec: corev1.PodSpec{
			HostPID:     true,
			HostNetwork: true,
			NodeSelector: map[string]string{
				"node-role.kubernetes.io/" + nodeRole: "",
			},
			Containers: []corev1.Container{
				{
					Name: "temp-container",
					Command: []string{
						"/bin/bash",
						"-c",
						"chroot /host systemctl status etcd",
						"sleep 10m",
					},
					Image: "busybox",
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
