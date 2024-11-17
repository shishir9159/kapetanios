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

	// TODO:
	//  refactor
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

	//curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.26/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
	//sudo chmod 644 /etc/apt/keyrings/kubernetes-apt-keyring.gpg # allow unprivileged APT programs to read this keyring
	//echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.26/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list
	//sudo chmod 644 /etc/apt/sources.list.d/kubernetes.list   # helps tools such as command-not-found to work correctly
	//apt update -y --allow-insecure-repositories
	//apt install -y kubectl=1.26.0-2.1 kubelet=1.26.0-2.1 kubeadm=1.26.0-2.1 --allow-unauthenticated
	//init --upload-certs
	//join --apiserver-advertise-address=<master-node-ip>

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
