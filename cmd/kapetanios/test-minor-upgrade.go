package main

import (
	"context"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/drain"
	"time"
)

func TestMinorUpgrade(namespace string) {

	logger, err := zap.NewProduction()
	defer func(logger *zap.Logger) {
		er := logger.Sync()
		if er != nil {
			logger.Fatal("error syncing logger before application terminates", zap.Error(err))
		}
	}(logger)

	// TODO:
	//  refactor
	client, err := orchestration.NewClient()

	// TODO: add namespace in the controller itself
	c := Controller{
		client: client,
		ctx:    context.TODO(),
		log:    logger,
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if err != nil {
		c.log.Error("error creating kubernetes client",
			zap.Error(err))
	}

	roleName := "minor-upgrade"

	renewalMinionManager := orchestration.NewMinions(client)

	nodes, err := c.client.Clientset().CoreV1().Nodes().Get(c.ctx, "shihab-node-1", metav1.GetOptions{})

	// TODO: sort with control-plane role, error no master nodes found

	if err != nil {
		c.log.Error("error listing nodes",
			zap.Error(err))
	}

	// TODO: refactor this part to orchestrator

	// namespace should only be included after the consideration for the existing
	// service account, cluster role binding
	descriptor := renewalMinionManager.MinionBlueprint("quay.io/klovercloud/minor-upgrade", roleName, nodes.Name)

	// TODO: instead of pod monitoring for creation, monitor for successful restarts
	//  er = RestartByLabel(c, map[string]string{"tier": "control-plane"}, node.Name)
	//  if er != nil {
	//  	c.log.Error("error restarting pods for certificate renewal",
	//	    	zap.Error(er))
	//	  //retry logic
	//	 // return er
	//	 break
	//  }

	// TODO: drain add node selector or something,
	//   add the same thing on the necessary pods(except for ds)

	//  TODO: after the pod is scheduled
	//   must first drain the node
	//   if failed, must be tainted again to
	//   schedule nodes

	descriptor.Spec.Tolerations = []corev1.Toleration{
		{
			Key:      "minor-upgrade-running",
			Operator: "Equal",
			Value:    "processing",
			// ---------

			Effect: corev1.TaintEffectNoSchedule,
		},
	}
	// ----------

	descriptor.Spec.HostNetwork = true

	// TODO: If any new Pods tolerate the node.kubernetes.io/unschedulable taint,
	//  then those Pods might be scheduled to the node you have drained.

	// TODO: force drain boolean

	c.log.Info("cordoning and draining node",
		zap.String("node name", nodes.Name))

	err = drainAndCordonNode(c, nodes)
	if err != nil {
		c.log.Error("failed to drain node",
			zap.String("node name:", nodes.Name),
			zap.Error(err))
	}

	// TODO:
	//  display the possible errors if force was not

	// TODO:
	//  if ran successfully
	//  Warning: deleting Pods that declare no controller: default/dnsutils; ignoring DaemonSet-managed Pods: kube-system/cilium-wnn6z, kube-system/kube-proxy-m8txw, metallb-system/speaker-587dh
	//evicting pod ingress-nginx/ingress-nginx-admission-create-5pjqv
	//evicting pod metallb-system/controller-9b6c9f6c9-g2p4z
	//evicting pod klovercloud/mesh-uat-go-two-6f846bfcf-ztnk5
	//evicting pod ingress-nginx/ingress-nginx-controller-5dcc84f655-pvdcc
	//evicting pod klovercloud/mesh-uat-go-one-86776497cf-227g9
	//evicting pod default/dnsutils
	//evicting pod default/backend-59b96df495-ghfn2
	//evicting pod ingress-nginx/ingress-nginx-admission-patch-4qgmn
	//evicting pod kube-system/coredns-787d4945fb-stk4w
	//evicting pod kube-system/cilium-operator-fdf6bc9f4-mcx4h
	//evicting pod default/kapetanios-b5bc457fb-v5vlx
	//evicting pod default/minions-for-etcd-4mg95
	//evicting pod kube-system/coredns-787d4945fb-4zl96
	//pod/ingress-nginx-admission-patch-4qgmn evicted
	//pod/ingress-nginx-admission-create-5pjqv evicted
	//pod/controller-9b6c9f6c9-g2p4z evicted
	//pod/minions-for-etcd-4mg95 evicted
	//pod/mesh-uat-go-one-86776497cf-227g9 evicted
	//I1029 03:31:47.241516 4025541 request.go:690] Waited for 1.103702573s due to client-side throttling, not priority and fairness, request: GET:https://10.0.0.3:6443/api/v1/namespaces/klovercloud/pods/mesh-uat-go-two-6f846bfcf-ztnk5
	//pod/dnsutils evicted
	//pod/backend-59b96df495-ghfn2 evicted
	//pod/kapetanios-b5bc457fb-v5vlx evicted
	//pod/mesh-uat-go-two-6f846bfcf-ztnk5 evicted
	//pod/cilium-operator-fdf6bc9f4-mcx4h evicted
	//pod/coredns-787d4945fb-4zl96 evicted
	//pod/coredns-787d4945fb-stk4w evicted
	//pod/ingress-nginx-controller-5dcc84f655-pvdcc evicted

	c.log.Info("tainting node",
		zap.String("node name", nodes.Name))

	addTaint(nodes)

	err = drain.RunCordonOrUncordon(&drain.Helper{
		Ctx:    c.ctx,
		Client: c.client.Clientset(),
	}, nodes, false)

	if err != nil {
		c.log.Error("error uncordoning the node",
			zap.String("node name", nodes.Name),
			zap.Error(err))
	}

	// TODO: refactor
	//drainer := &drain.Helper{
	//	Ctx:                             c.ctx,
	//	Client:                          c.client.Clientset(),
	//	DisableEviction:                 true,
	//	Force:                           true, // TODO: should it be Force eviction?
	//	IgnoreAllDaemonSets:             true,
	//	DeleteEmptyDirData:              true,
	//	SkipWaitForDeleteTimeoutSeconds: 30,
	//	Timeout:                         3 * time.Minute,
	//	GracePeriodSeconds:              10,
	//	Out:                             os.Stdout,
	//	ErrOut:                          os.Stderr,
	//}
	//
	//err = drain.RunCordonOrUncordon(drainer, nodes, false)

	// TODO:
	//  check for pods stuck in the terminating state
	//  if any pods other than the whitelisted ones are still in the node,
	//  force delete

	// TODO:
	//  if the pod doesn't schedule, check for taint
	//  check for all pod related event with informer

	// TODO: monitor the node status with watch

	// TODO: monitor the pod restart after upgrade
	//  All containers are restarted after upgrade, because the container spec hash value is changed
	//		just monitor the NODES before creating minion, no need to restart
	//  RestartByLabel(c, map[string]string{"tier": "control-plane"}, node.Name)

	minion, er := c.client.Clientset().CoreV1().Pods(namespace).Create(context.Background(), descriptor, metav1.CreateOptions{})
	if er != nil {
		c.log.Error("minor upgrade pod creation failed: ",
			zap.Error(er))

		//return er
		return
	}

	c.log.Info("minor upgrade pod created",
		zap.String("pod_name", minion.Name))

	time.Sleep(25 * time.Second)

	// TODO: All containers are restarted after upgrade, because the container spec hash value is changed.
	//   check if previously listed pods are all successfully restarted before untainted
	removeTaint(nodes)

	c.log.Info("after tainting")
}
