package main

import (
	"context"
	"flag"
	"github.com/rs/zerolog"
	pb "github.com/shishir9159/kapetanios/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"time"
)

var (
	addr = flag.String("addr", "kapetanios.default.svc.cluster.local:50051", "the address to connect to")
)

type Controller struct {
	ctx context.Context
	log zerolog.Logger
}

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// TODO: set log level

	logger := zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Timestamp().Logger()

	c := Controller{
		ctx: ctx,
		log: logger,
	}

	logger.Info().Msg("logging to os.Stdout")

	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		c.log.Error().Err(err).Msg("could not create grpc client")
	}

	//grpc.WithDisableServiceConfig()
	defer func(conn *grpc.ClientConn) {
		er := conn.Close()
		if er != nil {
			c.log.Error().Err(er).Msg("failed to close the grpc connection")
		}
	}(conn)

	connection := pb.NewValidityClient(conn)

	err = NodeHealth(c, connection)
	if err != nil {
		c.log.Error().Err(err).Msg("failed to get cluster health status")
	}

	expirationDate, daysRemaining, err := certExpiration(c, connection)

	if err != nil {
		c.log.Error().Err(err).Msg("failed to get cluster expiration date")
	}

	c.log.Info().
		Str("expirationDate", expirationDate.String()).
		Str("daysRemaining", daysRemaining.String()).
		Msg("checking certificate expiration date")
}
