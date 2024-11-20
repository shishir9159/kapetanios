package main

import (
	"context"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sync"
	"time"
)

type Controller struct {
	mu        sync.Mutex
	client    *orchestration.Client
	namespace string
	ctx       context.Context
	log       *zap.Logger
}

func Cert(namespace string) {

	logger := zap.Must(zap.NewProduction())
	defer func(logger *zap.Logger) {
		er := logger.Sync()
		if er != nil {
			logger.Fatal("error syncing logger before application terminates",
				zap.Error(er))
		}
	}(logger)

	client, err := orchestration.NewClient()

	c := Controller{
		client:    client,
		ctx:       context.Background(),
		log:       logger,
		namespace: namespace,
	}

	if err != nil {
		c.log.Error("error creating kubernetes client",
			zap.Error(err))
	}

	// TODO:
	//  Controller Definition need to be moved with the
	//  initial Setup and making sure there exists only one
	//  :refactor Controller name

	//for pkg in docker.io docker-doc docker-compose podman-docker containerd runc; do sudo apt-get remove $pkg; done
	//# Add Docker's official GPG key:
	//sudo apt-get update
	//sudo apt-get install ca-certificates curl
	//sudo install -m 0755 -d /etc/apt/keyrings
	//sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
	//sudo chmod a+r /etc/apt/keyrings/docker.asc
	//
	//# Add the repository to Apt sources:
	//echo \
	//  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
	//  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
	//  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
	//sudo apt-get update
	//sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin

	//curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.26/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
	//sudo chmod 644 /etc/apt/keyrings/kubernetes-apt-keyring.gpg # allow unprivileged APT programs to read this keyring
	//echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.26/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list
	//sudo chmod 644 /etc/apt/sources.list.d/kubernetes.list   # helps tools such as command-not-found to work correctly
	//apt update -y --allow-insecure-repositories
	//apt install -y kubectl=1.26.0-2.1 kubelet=1.26.0-2.1 kubeadm=1.26.0-2.1 --allow-unauthenticated
	//init --upload-certs --skip-phases=addon/kube-proxy
	//join --apiserver-advertise-address=<master-node-ip>

	//kubectl create secret generic -n kube-system cilium-etcd-secrets \
	//--from-file=etcd-client-ca.crt=/etc/kubernetes/pki/etcd-ca.pem \
	//--from-file=etcd-client.key=/etc/kubernetes/pki/etcd.key \
	//--from-file=etcd-client.crt=/etc/kubernetes/pki/etcd.cert

	// experiment with https://docs.cilium.io/en/stable/operations/performance/tuning/ given Supported NICs for BIG TCP: mlx4, mlx5, ice exists
	// check if --allocate-node-cidrs true in kube-controller-manager
	//API_SERVER_IP=10.0.0.7
	//helm template cilium/cilium --version 1.14.0 --namespace kube-system \
	//--set etcd.enabled=true --set etcd.ssl=true \
	//--set "etcd.endpoints[0]=https://10.0.0.7:2379" \
	//--set "etcd.endpoints[1]=https://10.0.0.9:2379" \
	//--set "etcd.endpoints[2]=https://10.0.0.10:2379" \
	//--set identityAllocationMode=kvstore \
	//--set kubeProxyReplacement=true \
	//--set bpf.hostLegacyRouting=false \
	//--set routingMode=native \
	//--set tunnelProtocol=geneve \
	//--set loadBalancer.dsrDispatch=geneve \
	//--set enable-ipv4-masquerade=true \
	//--set bpf.masquerade=true \
	//--set loadBalancer.mode=dsr \
	//--set enable-ipv6=false \
	//--set clean-cilium-bpf-state=true \
	//--set preallocate-bpf-maps=true \
	//--set cni.install=true \
	//--set cni.exclusive=true \
	//--set ipam.operator.clusterPoolIPv4PodCIDRList=10.244.0.0/16 \
	//--set ipam.mode=cluster-pool \
	//--set monitor-aggregation=true \
	//--set bpf.disableExternalIPMitigation=true \
	//--set loadBalancer.algorithm=maglev \
	//--set k8sServiceHost=${API_SERVER_IP} \
	//--set k8sServicePort=6443 \
	//--set externalTrafficPolicy=Local \
	//--set hubble.relay.enabled=true \
	//--set hubble.ui.enabled=true \
	//--output-dir manifests

	// validation: kubectl -n kube-system exec ds/cilium -- cilium-dbg status | grep KubeProxyReplacement
	// status: kubectl -n kube-system exec ds/cilium -- cilium-dbg status --verbose
	// status: kubectl -n kube-system exec ds/cilium -- cilium-dbg --all-addresses
	InitialSetup(c)

	roleName := "certs"
	matchLabels := map[string]string{"assigned-node-role-certs.kubernetes.io": roleName}

	labelSelector := metav1.LabelSelector{MatchLabels: matchLabels}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	renewalMinionManager := orchestration.NewMinions(client)

	nodes, err := c.client.Clientset().CoreV1().Nodes().List(context.Background(), listOptions)
	if err != nil {
		c.log.Error("error listing nodes",
			zap.Error(err))
	}

	if len(nodes.Items) == 0 {
		c.log.Error("no master nodes found",
			zap.Error(err))
		//	return err or call grpc
	}

	ch := make(chan *grpc.Server, 1)
	go CertGrpc(c.log, ch)

	// TODO: refactor this part to orchestrator by decomposing
	for index, node := range nodes.Items {

		// namespace should only be included after the consideration for the existing
		// service account, cluster role binding
		descriptor := renewalMinionManager.MinionBlueprint("quay.io/klovercloud/certs-renewal", roleName, node.Name)

		// kubectl get event --namespace default --field-selector involvedObject.name=minions
		// how many pods to monitor should be sent to the orchestration too
		minion, er := c.client.Clientset().CoreV1().Pods(namespace).Create(context.Background(), descriptor, metav1.CreateOptions{})
		if er != nil {
			c.log.Error("Cert Renewal pod creation failed: ",
				zap.Int("index", index),
				zap.Error(er))

			//return er
			return
		}

		c.log.Info("Cert Renewal pod created",
			zap.Int("index", index),
			zap.String("pod_name", minion.Name))

		// todo: wait for request for restart from the minions
		time.Sleep(3 * time.Second)

		er = RestartByLabel(c, map[string]string{"tier": "control-plane"}, node.Name)
		if er != nil {
			c.log.Error("error restarting pods for certificate renewal",
				zap.Error(er))

			//retry logic
			break
		}
	}

	(<-ch).GracefulStop()

	err = RestartRemainingComponents(c, "default")
	if err != nil {
		c.log.Error("error restarting renewal components",
			zap.Error(err))
	}
}
