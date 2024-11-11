package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	pb "github.com/shishir9159/kapetanios/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

var (
	addr = flag.String("addr", "kapetanios.default.svc.cluster.local:50051", "the address to connect to")
)

type Controller struct {
	ctx context.Context
	log *zap.Logger
}

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Println(err)
	}

	// replace zap with zeroLog

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
		log.Error("did not connect", zap.Error(err))
	}
	//grpc.WithDisableServiceConfig()
	defer func(conn *grpc.ClientConn) {
		er := conn.Close()
		if er != nil {
			log.Error("failed to close the grpc connection",
				zap.Error(er))
		}
	}(conn)

	connection := pb.NewRollbackClient(conn)

	err = prerequisites(c, connection)
	if err != nil {
		c.log.Error("the cluster didn't meet the condition for rollback",
			zap.Error(err))
	}

	//	step 1. replacing the last generated configs with
	//  the latest backups would cause the cert renewal rollback

	err = rollback(c, connection)
	if err != nil {
		c.log.Error("failed to renew certificates and kubeConfigs",
			zap.Error(err))
	}

	//step 3. Restarting pods to work with the updated certificates
	err = Restart(c, connection)
	if err != nil {
		c.log.Error("failed to restart kubernetes components after certificate renewal",
			zap.Error(err))
	}
}
