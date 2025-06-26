package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// if the labels exist
// if the labels match with the expectations
// number of control-plane pods and nodes running
// if the nodes match with the expectations

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

func validatingNodesState(nefario *Nefario, label string) error {

	//roleName := label

	//matchLabels := map[string]string{"assigned-node-role-" + label + "certs.kubernetes.io": roleName}
	//
	//labelSelector := metav1.LabelSelector{MatchLabels: matchLabels}
	//listOptions := metav1.ListOptions{
	//	LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	//}

	// TODO:

	nodes, err := nefario.client.Clientset().CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		nefario.log.Error("failed to list "+label+" nodes to validate",
			zap.Error(err))
		return err
	} else if nodes.Size() == 0 {
		nefario.log.Error("nodes for "+label+" are not labeled",
			zap.Error(err))
		return err
	}

	flag := true
	for _, node := range nodes.Items {
		if node.Status.Phase != corev1.NodeRunning {
			nefario.log.Error("node related to the operation is not running",
				zap.String("nodeName", node.Name))
			// todo:
			//  add to error and continue
			return fmt.Errorf("node %s is in the %s state", node.Name, node.Status.Phase)
		}

		if node.Labels["assigned-node-role-etcd.kubernetes.io"] == "true" {
			flag = false
		}
	}

	if flag {
		nefario.log.Error("no master nodes found")
		// TODO:
		//  error
		return fmt.Errorf("no master nodes found by querying with label")
	}

	return nil
}

func populatingConfigMap(nefario *Nefario) (*ETCD, error) {

	etcdCluster := ETCD{}

	cm, err := nefario.client.Clientset().CoreV1().ConfigMaps("kube-system").Get(context.Background(), "kubeadm-config", metav1.GetOptions{})

	if err != nil {
		nefario.log.Error("error fetching the kubeadm-config from the kube-system namespace",
			zap.Error(err))
		return &etcdCluster, err
	}

	// ClusterConfiguration stores the kubeadm-config as a file in the configmap
	yamlFile := cm.Data["ClusterConfiguration"]

	clusterConfiguration := ClusterConfiguration{}

	err = yaml.Unmarshal([]byte(yamlFile), &clusterConfiguration)
	if err != nil {
		nefario.log.Error("error parsing the kubeadm-config yaml file",
			zap.Error(err))
	}

	etcdCluster = clusterConfiguration.ETCD

	if etcdCluster.External.CAFile == "" {
		etcdCluster.External.CAFile = "/etc/kubernetes/pki/etcd-ca.pem"
	}

	if etcdCluster.External.CertFile == "" {
		etcdCluster.External.CertFile = "/etc/kubernetes/pki/etcd.cert"
	}

	if etcdCluster.External.KeyFile == "" {
		etcdCluster.External.KeyFile = "/etc/kubernetes/pki/etcd.key"
	}

	if len(etcdCluster.External.Endpoints) == 0 {
		return &etcdCluster, fmt.Errorf("no externl etcd endpoints provided")
	}

	configMapName := "kapetanios"

	configMap, er := nefario.client.Clientset().CoreV1().ConfigMaps(nefario.namespace).Get(context.Background(), configMapName, metav1.GetOptions{})
	if er != nil {
		nefario.log.Error("error fetching the configMap",
			zap.Error(er))
	}

	endpoints := strings.Join(etcdCluster.External.Endpoints, ";")

	configMap.Data["ETCD_NODES"] = endpoints
	configMap.Data["CERTIFICATES_DIR"] = removeTabsAndShiftWhitespaces(clusterConfiguration.CertificatesDir)
	configMap.Data["ETCD_CA_FILE"] = removeTabsAndShiftWhitespaces(clusterConfiguration.ETCD.External.CAFile)
	configMap.Data["ETCD_KEY_FILE"] = removeTabsAndShiftWhitespaces(clusterConfiguration.ETCD.External.KeyFile)
	configMap.Data["ETCD_CERT_FILE"] = removeTabsAndShiftWhitespaces(clusterConfiguration.ETCD.External.CertFile)
	configMap.Data["KUBERNETES_VERSION"] = removeTabsAndShiftWhitespaces(clusterConfiguration.KubernetesVersion)

	// todo: match by ip for etcd nodes
	for index, endpoint := range etcdCluster.External.Endpoints {
		configMap.Data["ETCD_NODE_"+strconv.Itoa(index+1)] = endpoint
	}

	_, er = nefario.client.Clientset().CoreV1().ConfigMaps(nefario.namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
	if er != nil {
		nefario.log.Error("error updating configMap",
			zap.Error(er))
	}

	return &etcdCluster, nil
}

func continueMinorUpgrade(upgrade *MinorUpgrade) error {
	upgrade.nefario.log.Info("upgrade is continued after the successful restart",
		zap.Strings("nodes to be upgraded", upgrade.upgradeState.NodesStatus.NodesToBeUpgraded))

	upgrade.finalizer = make(chan bool, 1)

	upgrade.mu.Lock()
	go upgrade.MinorUpgrade()

	return nil
}

func prerequisites(upgrade *MinorUpgrade) error {

	etcd, err := populatingConfigMap(upgrade.nefario)
	if err != nil {
		fmt.Println("error populating configMap")
		return err
	}
	fmt.Println("etcd: ", etcd)

	// if configMap shows updated nodes to a certain value
	// and a desired k8s version must exist on the cm
	// after the updates

	// TODO: throw error no master nodes found

	validatingNodesState(upgrade.nefario, "control-plane")

	// TODO:
	// 	move clean up pods in internal library

	//labelSelector := metav1.LabelSelector{
	//	MatchLabels: map[string]string{
	//		"app": "minor-upgrade",
	//		"app": "certs",
	//	},
	//}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	listOptions := metav1.ListOptions{
		LabelSelector: "app in (certs, minor-upgrade), app notin (etcd, kubelet, control-plane)",
		//LabelSelector: "app in (certs, minor-upgrade), app notin (etcd, kubelet, control-plane)",
		//LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	minions, err := upgrade.nefario.client.Clientset().CoreV1().Pods(os.Getenv("KAPETANIOS_NAMESPACE")).
		List(ctx, listOptions)
	if err != nil {
		upgrade.nefario.log.Error("error listing completed/failed pods for node management",
			zap.Error(err))
		return err
	}

	for _, minion := range minions.Items {
		er := upgrade.nefario.client.Clientset().CoreV1().Pods(os.Getenv("KAPETANIOS_NAMESPACE")).
			Delete(ctx, minion.Name, metav1.DeleteOptions{})
		if er != nil {
			upgrade.nefario.log.Info("failed to delete minion:",
				zap.String("minion name:", minion.Name),
				zap.Error(er))
		}
	}

	// TODO:
	//  decide what to do
	nodes, err := upgrade.nefario.client.Clientset().CoreV1().Nodes().
		List(ctx, listOptions)
	if err != nil {
		upgrade.nefario.log.Error("error listing completed/failed pods for node management",
			zap.Error(err))
	}

	flag := true
	for _, node := range nodes.Items {
		if node.Labels["assigned-node-role-certs.kubernetes.io"] == "etcd" {
			flag = false
			break
		}
	}

	if flag {

	}

	// TODO: race condition - readCtx can be cancelled

	return nil
}
