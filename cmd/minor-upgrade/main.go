package main

import (
	"context"
	"flag"
	"fmt"
	pb "github.com/shishir9159/kapetanios/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"time"
	//"github.com/rs/zerolog/log"
)

var (
	addr = flag.String("addr", "kapetanios.default.svc.cluster.local:50051", "the address to connect to")
)

type Controller struct {
	ctx context.Context
	log *zap.Logger
}

func GrpcClient(log *zap.Logger) {

	var addr *string

	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, network, "10.96.0.10:53")
		},
	}

	ips, err := r.LookupNetIP(context.Background(), "ip4", "kapetanios.default.svc.cluster.local")
	log.Info("kapetanios service address")
	fmt.Println(ips)

	if len(ips) != 0 {
		for i, ip := range ips {
			if ip.BitLen() == 32 {
				log.Info("kapetanios service address", zap.String("ip", ip.String()))
				addr = flag.String("addr", ip.String()+":50051", "the address to connect to")
				log.Info("address", zap.Int("index", i), zap.String("addr", *addr))
			}
		}
	}

	if err != nil {
		log.Error("error kapetanios address", zap.Error(err))
	}

	if addr == nil {
		log.Error("addr was empty")
		addr = flag.String("addr", "kapetanios.default.svc.cluster.local:50051", "the address to connect to")
	}

	flag.Parse()

	log.Info("connecting to kapetanios service address", zap.String("addr", *addr))

	log.Info("1")
	// Set up a connection to the server.
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("did not connect", zap.Error(err))
	}
	log.Info("2")
	//grpc.WithDisableServiceConfig()
	defer func(conn *grpc.ClientConn) {
		er := conn.Close()
		if er != nil {
			log.Error("failed to close the grpc connection",
				zap.Error(er))
		}
	}(conn)

	log.Info("3")
	c := pb.NewMinorUpgradeClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	log.Info("4")
	defer cancel()

	rpc, err := c.ClusterHealthChecking(ctx,
		&pb.PrerequisitesMinorUpgrade{
			EtcdStatus:          false,
			StorageAvailability: 0,
			Err:                 "",
		})

	if err != nil {
		log.Error("could not send status update: ", zap.Error(err))
	}

	log.Info("Backup Status",
		zap.Bool("next step", rpc.GetProceedNextStep()),
		zap.Bool("retry", rpc.GetSkipRetryCurrentStep()),
		zap.Bool("terminate application", rpc.GetTerminateApplication()))
}

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Println(err)
	}

	// TODO:
	//  replace zap with zeroLog

	c := Controller{
		ctx: ctx,
		log: logger,
	}

	defer func(logger *zap.Logger) {
		er := logger.Sync()
		if er != nil {
			c.log.Error("failed to close logger",
				zap.Error(er))
		}
	}(logger)

	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		c.log.Error("did not connect", zap.Error(err))
	}
	//grpc.WithDisableServiceConfig()
	defer func(conn *grpc.ClientConn) {
		er := conn.Close()
		if er != nil {
			c.log.Error("failed to close the grpc connection",
				zap.Error(er))
		}
	}(conn)

	err = Prerequisites(c, conn)
	if err != nil {

	}

	availableVersionList, err := availableVersions(c, conn)

	if len(availableVersionList) == 0 {
		c.log.Fatal("no available versions for minor upgrade",
			zap.Error(err))
	}

	if err != nil {
		c.log.Error("failed to fetch minor versions for kubernetes version upgrade",
			zap.Error(err))
	}

	// todo: include in the testing
	testing := false
	latest := false
	version := "1.26.6-1.1"

	if testing {
		//version = kubernetesVersion
	}

	if latest {
	}

	// TODO:
	//  if available version fails or works,
	//  check if that matches with upgradePlane
	//  if the latest is selected

	// TODO: refactor
	plan := "v1.26.6"

	diff, err := compatibility(c, plan, conn)
	if err != nil {
		c.log.Error("failed to get diff",
			zap.Error(err))
	}

	c.log.Info("diff for upgrade plan",
		zap.String("diff", diff))

	_, err = k8sComponentsUpgrade(c, "kubeadm", version, conn)
	if err != nil {
		c.log.Error("failed to get upgrade kubeadm",
			zap.Error(err))

	}

	_, err = upgradePlan(c, conn)
	if err != nil {
		c.log.Error("failed to get upgrade plan",
			zap.Error(err))
	}

	_, err = clusterUpgrade(c, version, conn)
	if err != nil {
		c.log.Error("failed to get upgrade plan",
			zap.Error(err))
	}

	_, err = k8sComponentsUpgrade(c, "kubelet", version, conn)
	if err != nil {
		c.log.Error("failed to get upgrade plan",
			zap.Error(err))
	}

	err = restartComponent(c, "kubelet", conn)
	if err != nil {
		c.log.Error("failed to restart kubelet",
			zap.Error(err))
	}

	// TODO: etcd node restart based
	//  on condition

	_, err = k8sComponentsUpgrade(c, "kubectl", version, conn)
	if err != nil {
		c.log.Error("failed to get upgrade plan",
			zap.Error(err))
	}

	// TODO: sanityChecking & finalizer
}
