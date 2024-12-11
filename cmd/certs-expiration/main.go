package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/rs/zerolog"
	pb "github.com/shishir9159/kapetanios/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	addr        = flag.String("addr", "kapetanios.default.svc.cluster.local:50051", "the address to connect to")
	retryPolicy = `{
		"methodConfig": [{
		  "retryPolicy": {
			  "MaxAttempts": 4,
			  "InitialBackoff": ".01s",
			  "MaxBackoff": ".01s",
			  "BackoffMultiplier": 4.0,
			  "RetryableStatusCodes": [ "UNAVAILABLE", "DEADLINE_EXCEEDED" ]
		  }
		}]}`
)

type Controller struct {
	ctx context.Context
	log zerolog.Logger
}

// TODO: parallel execution
func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// TODO: set log level

	logger := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339Nano,
		FormatLevel: func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("[%s]", i))
		},
		FormatMessage: func(i interface{}) string {
			return fmt.Sprintf("| %s |", i)
		},
		FormatCaller: func(i interface{}) string {
			return filepath.Base(fmt.Sprintf("%s", i))
		},
		PartsExclude: []string{
			zerolog.TimestampFieldName,
		},
	}).With().Timestamp().Caller().Stack().Logger()
	//.Level(zerolog.InfoLevel)

	c := Controller{
		ctx: ctx,
		log: logger,
	}

	c.log.Info().Msg("logging to os.Stdout")

	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(retryPolicy))
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
