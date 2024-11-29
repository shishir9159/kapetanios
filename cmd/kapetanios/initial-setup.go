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
	"strconv"
	"strings"
)

type ETCD struct {
	External struct {
		Endpoints []string `yaml:"endpoints"`
		CAFile    string   `yaml:"caFile"`
		CertFile  string   `yaml:"certFile"`
		KeyFile   string   `yaml:"keyFile"`
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

// TODO:
//  fetch localAPIEndpoint: advertiseAddress

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
		c.log.Error("failed to list "+label+" nodes to validate",
			zap.Error(err))
		return err
	} else if nodes.Size() == 0 {
		c.log.Error("nodes for "+label+" are not labeled",
			zap.Error(err))
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

func populatingConfigMap(c Controller) (*ETCD, error) {

	etcdCluster := ETCD{}

	cm, err := c.client.Clientset().CoreV1().ConfigMaps("kube-system").Get(context.Background(), "kubeadm-config", metav1.GetOptions{})

	if err != nil {
		c.log.Error("error fetching the kubeadm-config from the kube-system namespace",
			zap.Error(err))
		return &etcdCluster, err
	}

	// ClusterConfiguration stores the kubeadm-config as a file in the configmap
	yamlFile := cm.Data["ClusterConfiguration"]

	clusterConfiguration := ClusterConfiguration{}

	err = yaml.Unmarshal([]byte(yamlFile), &clusterConfiguration)
	if err != nil {
		c.log.Error("error parsing the kubeadm-config yaml file",
			zap.Error(err))
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

	if len(etcdCluster.External.Endpoints) == 0 {
		return &etcdCluster, fmt.Errorf("no externl etcd endpoints provided")
	}

	configMapName := "kapetanios"

	configMap, er := c.client.Clientset().CoreV1().ConfigMaps(c.namespace).Get(context.Background(), configMapName, metav1.GetOptions{})
	if er != nil {
		c.log.Error("error fetching the configMap",
			zap.Error(er))
	}

	endpoints := strings.Join(etcdCluster.External.Endpoints, ";")

	configMap.Data["ETCD_NODES"] = endpoints
	configMap.Data["KUBERNETES_VERSION"] = removeTabsAndShiftWhitespaces(clusterConfiguration.KubernetesVersion)
	configMap.Data["CertificatesDir"] = removeTabsAndShiftWhitespaces(clusterConfiguration.CertificatesDir)
	configMap.Data["ETCD_CA_FILE"] = removeTabsAndShiftWhitespaces(clusterConfiguration.ETCD.External.CAFile)
	configMap.Data["ETCD_CERT_FILE"] = removeTabsAndShiftWhitespaces(clusterConfiguration.ETCD.External.CertFile)
	configMap.Data["ETCD_KEY_FILE"] = removeTabsAndShiftWhitespaces(clusterConfiguration.ETCD.External.KeyFile)

	// todo: match by ip for etcd nodes
	for index, endpoint := range etcdCluster.External.Endpoints {
		configMap.Data["ETCD_NODE_"+strconv.Itoa(index+1)] = endpoint
	}

	_, er = c.client.Clientset().CoreV1().ConfigMaps(c.namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
	if er != nil {
		c.log.Error("error updating configMap",
			zap.Error(er))
	}

	return &etcdCluster, nil
}

func InitialSetup(c Controller) {

	etcdCluster, err := populatingConfigMap(c)
	if err != nil {
		c.log.Error("error populating config",
			zap.Error(err))
	}

	// TODO:
	//  return the etcdCluster
	// TODO:
	//  create a cache layer
	fmt.Println(etcdCluster)

	err = validatingNodesState(c, "certs")
	if err != nil {
		c.log.Error("error validating master node labels",
			zap.Error(err))
	}

	err = validatingNodesState(c, "etcd")
	if err != nil {
		c.log.Error("error validating etcd node labels",
			zap.Error(err))
	}

}
