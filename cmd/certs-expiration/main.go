package main

import (
	"context"
	"flag"
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	logger := zap.Must(zap.NewProduction())

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
		log.Error("did not connect",
			zap.Error(err))
	}

	//grpc.WithDisableServiceConfig()
	defer func(conn *grpc.ClientConn) {
		er := conn.Close()
		if er != nil {
			log.Error("failed to close the grpc connection",
				zap.Error(er))
		}
	}(conn)

	connection := pb.NewValidityClient(conn)

	err = NodeHealth(c, connection)
	if err != nil {
		c.log.Error("failed to get cluster health status",
			zap.Error(err))
	}

	expirationDate, daysRemaining, err := certExpiration(c, connection)
	c.log.Info("checking certificate expiration date",
		zap.String("expirationDate", expirationDate.String()),
		zap.String("daysRemaining", daysRemaining.String()))

}
