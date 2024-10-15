package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"os"
	"regexp"
)

type ETCD struct {
	External struct {
		Endpoints []string `json:"endpoints"`
		CAFile    string   `json:"caFile"`
		CertFile  string   `json:"certFile"`
		KeyFile   string   `json:"keyFile"`
	} `yaml:"external"`
}

type ClusterConfiguration struct {
	ApiServer struct {
		ExtraArgs []struct {
			Arg string `yaml:"arg"`
		}
		ExtraVolumes []struct {
			HostPath  string `yaml:"hostPath"`
			MountPath string `yaml:"mountPath"`
			Name      string `yaml:"name"`
			ReadOnly  string `yaml:"readOnly"`
		} `yaml:"extraVolumes"`
		TimeoutForControlPlane string `yaml:"timeoutForControlPlane"`
	} `yaml:"apiServer"`
	ApiVersion           string            `yaml:"apiVersion"`
	CertificatesDir      string            `yaml:"certificatesDir"`
	ClusterName          string            `yaml:"clusterName"`
	ControlPlaneEndpoint string            `yaml:"controlPlaneEndpoint"`
	ControllerManager    map[string]string `yaml:"controllerManager"`
	DNS                  map[string]string `yaml:"dns"`
	ETCD                 ETCD              `yaml:"etcd"`
	ImageRepository      string            `yaml:"imageRepository"`
	Kind                 string            `yaml:"kind"`
	KubernetesVersion    string            `yaml:"kubernetesVersion"`
	Networking           struct {
		DnsDomains    string `yaml:"dnsDomains"`
		ServiceSubnet string `yaml:"serviceSubnet"`
	} `yaml:"networking"`
	Scheduler map[string]string `yaml:"scheduler"`
}

//    apiServer:
//      extraArgs:
//        audit-log-maxage: "7"
//        audit-log-path: /var/log/k8_audit.log
//        audit-policy-file: /etc/kubernetes/audit-policy.yaml
//        authorization-mode: Node,RBAC
//        enable-admission-plugins: ResourceQuota,AlwaysPullImages,DefaultStorageClass
//        max-mutating-requests-inflight: "500"
//        max-requests-inflight: "2000"
//      extraVolumes:
//      - hostPath: /etc/kubernetes/audit-policy.yaml
//        mountPath: /etc/kubernetes/audit-policy.yaml
//        name: audit-policy
//        readOnly: true
//      - hostPath: /var/log/
//        mountPath: /var/log/
//        name: audit-log
//      timeoutForControlPlane: 4m0s
//    apiVersion: kubeadm.k8s.io/v1beta3
//    certificatesDir: /etc/kubernetes/pki
//    clusterName: kubernetes
//    controlPlaneEndpoint: 10.0.0.3:6443
//    controllerManager: {}
//    dns: {}
//    etcd:
//      external:
//        caFile: /etc/kubernetes/pki/etcd-ca.pem
//        certFile: /etc/kubernetes/pki/etcd.cert
//        endpoints:
//        - https://5.161.64.103:2379
//        - https://5.161.248.112:2379
//        - https://5.161.67.249:2379
//        keyFile: /etc/kubernetes/pki/etcd.key
//    imageRepository: registry.k8s.io
//    kind: ClusterConfiguration
//    kubernetesVersion: v1.26.15
//    networking:
//      dnsDomain: cluster.local
//      serviceSubnet: 10.96.0.0/12
//    scheduler: {}

//root@shihab-node-1:~/kapetanios# cat /etc/kubernetes/kubeadm-config.yaml
//apiVersion: kubeadm.k8s.io/v1beta2
//kind: InitConfiguration
//nodeRegistration:
//  criSocket: "unix:///run/containerd/containerd.sock"
//localAPIEndpoint:
//  advertiseAddress: 10.0.0.3
//  bindPort: 6443
//
//---
//apiVersion: kubeadm.k8s.io/v1beta2
//kind: ClusterConfiguration
//kubernetesVersion: stable
//controlPlaneEndpoint: 10.0.0.3
//apiServer:
//  extraArgs:
//    enable-admission-plugins:  ResourceQuota,AlwaysPullImages,DefaultStorageClass
//    max-mutating-requests-inflight: "500"
//    max-requests-inflight: "2000"
//    audit-log-path: /var/log/k8_audit.log
//    audit-policy-file: /etc/kubernetes/audit-policy.yaml
//    audit-log-maxage: "7"
//  extraVolumes:
//    - name: audit-policy
//      hostPath: /etc/kubernetes/audit-policy.yaml
//      mountPath: /etc/kubernetes/audit-policy.yaml
//      readOnly: true
//    - name: audit-log
//      hostPath: /var/log/
//      mountPath: /var/log/
//      readOnly: false
//etcd:
//  external:
//     endpoints:
//       - https://5.161.64.103:2379
//       - https://5.161.248.112:2379
//       - https://5.161.67.249:2379
//     caFile: /etc/kubernetes/pki/etcd-ca.pem
//     certFile: /etc/kubernetes/pki/etcd.cert
//     keyFile: /etc/kubernetes/pki/etcd.key

//---------------------------------------------------

// TODO:
//  fetch localAPIEndpoint: advertiseAddress
//  store certificate validity

func removeTabsAndShiftWhitespaces(s string) string {

	// Regular expression to match tabs and shift whitespaces
	re := regexp.MustCompile(`[\x00-\x1F\x7F\t\s<\nil>]+`)
	return re.ReplaceAllString(s, "")

}

func validatingNodesState(c Controller, label string) error {

	roleName := label
	matchLabels := map[string]string{"assigned-node-role-" + label + "certs.kubernetes.io": roleName}

	labelSelector := metav1.LabelSelector{MatchLabels: matchLabels}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	nodes, err := c.client.Clientset().CoreV1().Nodes().List(context.Background(), listOptions)
	if err != nil {
		c.log.Error("failed to list "+label+" nodes to validate", zap.Error(err))
		return err
	} else if nodes.Size() == 0 {
		c.log.Error("nodes for "+label+" are not labeled", zap.Error(err))
		return err
	}

	for _, node := range nodes.Items {
		if node.Status.Phase != corev1.NodeRunning {
			c.log.Error("node related to the operation is not running",
				zap.String("nodeName", node.Name))
			return fmt.Errorf("node %s is in the %s state", node.Name, node.Status.Phase)
		}
	}

	return nil
}

// TODO:
//  refactor this duplicated code to the internal

func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

// TODO:
//  save certDir to configMap

func populatingConfigMap(c Controller) (*ETCD, error) {

	etcdCluster := ETCD{}

	cm, err := c.client.Clientset().CoreV1().ConfigMaps("kube-system").Get(context.Background(), "kubeadm-config", metav1.GetOptions{})

	if err != nil {
		c.log.Error("error fetching the kubeadm-config from the kube-system namespace", zap.Error(err))
		return &etcdCluster, err
	}

	// ClusterConfiguration stores the kubeadm-config as a file in the configmap

	yamlFile := cm.Data["ClusterConfiguration"]

	clusterConfiguration := ClusterConfiguration{}

	err = yaml.Unmarshal([]byte(yamlFile), &clusterConfiguration)
	if err != nil {
		c.log.Error("error parsing the kubeadm-config yaml file", zap.Error(err))
	}

	etcdCluster = clusterConfiguration.ETCD

	if etcdCluster.External.CAFile == "" {
		//	TODO:
		//	 check if the file exists
		//   throw permission errors
		if Exists("/etc/kubernetes/pki/etcd-ca.pem") {
			return &etcdCluster, fmt.Errorf("etcd ca doesn't exist or read permission error")
		}
	}

	if etcdCluster.External.CertFile == "" {
		//	TODO:
		//	 check if the file exists
		//   throw permission errors
		if Exists("/etc/kubernetes/pki/etcd.cert") {
			return &etcdCluster, fmt.Errorf("etcd cert doesn't exist or read permission error")
		}
	}

	if etcdCluster.External.KeyFile == "" {
		//	TODO:
		//	 check if the file exists
		//   throw permission errors
		if Exists("/etc/kubernetes/pki/etcd.key") {
			return &etcdCluster, fmt.Errorf("etcd key file doesn't exist or read permission error")
		}
	}

	c.log.Info("", zap.String("kubernetesVersion", removeTabsAndShiftWhitespaces(clusterConfiguration.KubernetesVersion)))
	c.log.Info("", zap.String("caFile", removeTabsAndShiftWhitespaces(clusterConfiguration.ETCD.External.CAFile)))
	c.log.Info("", zap.String("certFile", removeTabsAndShiftWhitespaces(clusterConfiguration.ETCD.External.CertFile)))
	c.log.Info("", zap.String("keyFile", removeTabsAndShiftWhitespaces(clusterConfiguration.ETCD.External.KeyFile)))

	for index, endpoint := range clusterConfiguration.ETCD.External.Endpoints {
		c.log.Info("", zap.Int("index", index),
			zap.String("endpoint", removeTabsAndShiftWhitespaces(endpoint)))
	}

	fmt.Println("check 1 ", clusterConfiguration.KubernetesVersion)
	fmt.Println("check 2 ", clusterConfiguration.ETCD.External.CAFile)
	fmt.Printf("check 3 %s", clusterConfiguration.ETCD.External.Endpoints[0])

	return &etcdCluster, nil
}

func InitialSetup(c Controller) {

	etcdCluster, err := populatingConfigMap(c)
	if err != nil {
		c.log.Error("error populating config", zap.Error(err))
	}

	// TODO:
	//  return the etcdCluster
	fmt.Println(etcdCluster)

	err = validatingNodesState(c, "certs")
	if err != nil {
		c.log.Error("error validating master node labels", zap.Error(err))
	}

	err = validatingNodesState(c, "etcd")
	if err != nil {
		c.log.Error("error validating etcd node labels", zap.Error(err))
	}

}
