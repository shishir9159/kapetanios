package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	maxAttempts = 3
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
	ctx    context.Context
	log    zerolog.Logger
	distro string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	//if debug == true {
	//	out := os.Stdout
	//	logLevel := zerolog.DebugLevel
	//}

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

	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(retryPolicy))
	if err != nil {
		c.log.Error().Err(err).
			Msg("couldn't connect to the kapetanios")

	}

	//grpc.WithDisableServiceConfig()
	defer func(conn *grpc.ClientConn) {
		er := conn.Close()
		if er != nil {
			c.log.Error().Err(er).
				Msg("failed to close the grpc connection")
		}
	}(conn)

	for i := 0; i < maxAttempts; i++ {
		var skip bool
		skip, c.distro, err = Prerequisites(c, conn)
		if err != nil {
			c.log.Error().Err(err).
				Int("attempt", i).
				Msg("failed to check prerequisites for cluster upgrade")
		}

		if skip {
			break
		}
	}

	if c.distro == "" {
		c.log.Fatal().Err(err).
			Msg("couldn't fetch distro info")
	}

	for i := 0; i < maxAttempts; i++ {
		var skip bool
		skip, err = availableVersions(c, conn)

		if err != nil {
			c.log.Error().Err(err).
				Int("attempt", i).
				Msg("failed to fetch minor versions for the kubernetes upgrade")
		}

		if skip {
			break
		}
	}

	if os.Getenv("DEBUG") == "true" {
		//TODO: version = current kubernetesVersion
	}
}
