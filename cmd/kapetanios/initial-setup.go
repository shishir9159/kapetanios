package main

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
)

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
	ETCD                 struct {
		External struct {
			CAFile    string   `json:"caFile"`
			CertFile  string   `json:"certFile"`
			Endpoints []string `json:"endpoints"`
			KeyFile   string   `json:"keyFile"`
		} `yaml:"external"`
	} `yaml:"etcd"`
	ImageRepository   string `yaml:"imageRepository"`
	Kind              string `yaml:"kind"`
	KubernetesVersion string `yaml:"kubernetesVersion"`
	Networking        struct {
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

//---

// store certificate validity
// check number of nodes
// save certDir to configMap
// checking if the necessary files exist

func removeTabsAndShiftWhitespaces(s string) string {
	// Regular expression to match tabs and shift whitespaces
	re := regexp.MustCompile(`[\t\s]+`)

	// Replace matched characters with an empty string
	return re.ReplaceAllString(s, "")
}

func populatingConfigMap(c Controller) error {

	cm, err := c.client.Clientset().CoreV1().ConfigMaps("kube-system").Get(context.Background(), "kubeadm-config", metav1.GetOptions{})

	if err != nil {
		c.log.Error("error fetching the kubeadm-config from the kube-system namespace", zap.Error(err))
		return err
	}

	// ClusterConfiguration stores the kubeadm-config as a file in the configmap

	yamlFile := cm.Data["ClusterConfiguration"]

	clusterConfiguration := ClusterConfiguration{}

	err = yaml.Unmarshal([]byte(yamlFile), &clusterConfiguration)
	if err != nil {
		log.Error("error parsing the kubeadm-config yaml file", zap.Error(err))
	}

	fmt.Println(clusterConfiguration)

	log.Info(zap.String("kubernetesVersion", removeTabsAndShiftWhitespaces(clusterConfiguration.KubernetesVersion)))
	log.Info(zap.String("caFile", removeTabsAndShiftWhitespaces(clusterConfiguration.ETCD.External.CAFile)))
	log.Info(zap.String("certFile", removeTabsAndShiftWhitespaces(clusterConfiguration.ETCD.External.CertFile)))
	log.Info(zap.String("keyFile", removeTabsAndShiftWhitespaces(clusterConfiguration.ETCD.External.KeyFile)))

	for index, endpoint := range clusterConfiguration.ETCD.External.Endpoints {
		log.Info(zap.Int("index", index),
			zap.String("endpoint", removeTabsAndShiftWhitespaces(endpoint)))
	}
	fmt.Println("check 1 ", clusterConfiguration.KubernetesVersion)
	fmt.Println("check 2 ", clusterConfiguration.ETCD.External.CAFile)
	fmt.Printf("check 3 %s", clusterConfiguration.ETCD.External.Endpoints[0])

	return nil
}

func InitialSetup(c Controller) {

	err := populatingConfigMap(c)
	if err != nil {
		c.log.Error("error populating config", zap.Error(err))
	}
}
