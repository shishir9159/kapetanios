package main

import (
	"context"
	"fmt"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/drain"
	"os"
	"sort"
	"sync"
	"time"
)

func drainAndCordonNode(c Controller, node *corev1.Node) error {

	drainer := &drain.Helper{
		Client:                          c.client.Clientset(),
		DisableEviction:                 true,
		Force:                           true, // TODO: should it be Force eviction?
		IgnoreAllDaemonSets:             true,
		DeleteEmptyDirData:              true,
		SkipWaitForDeleteTimeoutSeconds: 30,
		Timeout:                         2 * time.Minute,
		GracePeriodSeconds:              10,
		Out:                             os.Stdout,
		ErrOut:                          os.Stderr,
	}

	err := drain.RunCordonOrUncordon(drainer, node, true)
	if err != nil {

	}

	err = drain.RunNodeDrain(drainer, node.Name)
	if err != nil {

	}

	return nil
}

func removeTaint(node *corev1.Node) {

	taints := node.Spec.Taints

	if len(taints) == 0 {
		return
	}

	taintToRemove := corev1.Taint{
		Key:    "minor-upgrade-running",
		Value:  "processing",
		Effect: corev1.TaintEffectNoSchedule,
	}

	newTaints := []corev1.Taint{taintToRemove}

	for _, taint := range taints {
		if taint.MatchTaint(&taintToRemove) {
			continue
		}

		newTaints = append(newTaints, taint)
	}

	node.Spec.Taints = newTaints
}

func addTaint(node *corev1.Node) {

	taints := node.Spec.Taints

	// TODO: declare as a struct maybe?
	taintToAdd := corev1.Taint{
		Key:    "minor-upgrade-running",
		Value:  "processing",
		Effect: corev1.TaintEffectNoSchedule,
	}

	newTaints := []corev1.Taint{taintToAdd}

	if len(taints) != 0 {
		for _, taint := range taints {
			if taint.MatchTaint(&taintToAdd) {
				return
			}

			newTaints = append(newTaints, taint)
		}

		return
	}

	node.Spec.Taints = newTaints
}

// be careful about the different  across
// the nodes version

// TODO: https://v1-27.docs.kubernetes.io/docs/tasks/administer-cluster/kubeadm/kubeadm-upgrade/#recovering-from-a-failure-state
func recovery(namespace string) {

}

func next() {
		current := util.MustParseSemantic(current)
		target := util.MustParseSemantic(desired)
		var nextMinor uint
		if target.Minor() == current.Minor() {
			nextMinor = current.Minor()
		} else {
			nextMinor = current.Minor() + 1
		}

		if nextMinor == target.Minor() {
			if _, ok := files.FileSha256["kubeadm"]["amd64"][desired]; !ok {
				return "", errors.Errorf("the target  %s is not supported", desired)
			}
			return desired, nil
		} else {
			nextPatchList := make([]int, 0)
			for supportStr := range files.FileSha256["kubeadm"]["amd64"] {
				support := util.MustParseSemantic(supportStr)
				if support.Minor() == nextMinor {
					nextPatchList = append(nextPatchList, int(support.Patch()))
				}
			}
			sort.Ints(nextPatchList)

			next := current.WithMinor(nextMinor)
			if len(nextPatchList) == 0 {
				return "", errors.Errorf("Kubernetes minor  v%d.%d.x is not supported", next.Major(), next.Minor())
			}
			next = next.WithPatch(uint(nextPatchList[len(nextPatchList)-1]))

			return fmt.Sprintf("v%s", next.String()), nil
		}
	}
}

// TODO: for testing purposes, try the current 

func MinorUpgrade(namespace string) {

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

	// TODO: add mutex
	// TODO: add namespace in the controller itself
	c := Controller{
		client: client,
		ctx:    context.Background(),
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

	nodes, err := c.client.Clientset().CoreV1().Nodes().List(context.Background(), metav1.ListOptions{LabelSelector: ""})

	// TODO: sort with control-plane role, error no master nodes found

	if err != nil {
		c.log.Error("error listing nodes",
			zap.Error(err))
	}

	if len(nodes.Items) == 0 {
		c.log.Error("no nodes found",
			//	return err or call grpc
			zap.Error(err))
	}

	// TODO: refactor this part to orchestrator

	for index, node := range nodes.Items {

		// namespace should only be included after the consideration for the existing
		// service account, cluster role binding
		descriptor := renewalMinionManager.MinionBlueprint("quay.io/klovercloud/minor-upgrade", roleName, node.Name)

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
				Key:               "minor-upgrade-running",
				Operator:          "Equal",
				Value:             "processing",
				Effect:            "NoSchedule",
				TolerationSeconds: &[]int64{3}[0],
			},
		}

		// TODO: If any new Pods tolerate the node.kubernetes.io/unschedulable taint,
		//  then those Pods might be scheduled to the node you have drained.

		err = drainAndCordonNode(c, &node)
		if err != nil {
			c.log.Error("failed to drain node",
				zap.String("node name:", node.Name),
				zap.Error(err))
		}

		addTaint(&node)

		err = drain.RunCordonOrUncordon(&drain.Helper{}, &node, false)

		// TODO:
		//  if the pod doesn't schedule, check for taint
		//  check for all pod related event with informer

		// TODO: monitor the node status with watch

		// TODO: monitor the pod restart after upgrade
		//  All containers are restarted after upgrade, because the container spec hash value is changed
		//		just monitor the NODES before creating minion, no need to restart
		//RestartByLabel(c, map[string]string{"tier": "control-plane"}, node.Name)

		minion, er := c.client.Clientset().CoreV1().Pods(namespace).Create(context.Background(), descriptor, metav1.CreateOptions{})
		if er != nil {
			c.log.Error("minor upgrade pod creation failed: ",
				zap.Int("index", index),
				zap.Error(er))

			//return er
			return
		}

		c.log.Info("minor upgrade pod created",
			zap.Int("index", index),
			zap.String("pod_name", minion.Name))

		// todo: wait for request for restart from the minions
		time.Sleep(5 * time.Second)

		// TODO: All containers are restarted after upgrade, because the container spec hash value is changed.
		//   check if previously listed pods are all successfully restarted before untainted
		removeTaint(&node)
	}
}

//  ToDo:
//   update the information in the configMaps
//   specially about k8s version
