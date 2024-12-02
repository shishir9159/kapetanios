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
	ctx context.Context
	log zerolog.Logger
}

func main() {

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
		ctx: context.Background(),
		log: logger,
	}

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(retryPolicy))
	if err != nil {
		c.log.Error().Err(err).
			Msg("couldn't connect to the kapetanios")
	}

	fmt.Println(conn)

	//	TODO: etcd remove
	//	 ETCDCTL_API=3 etcdctl endpoint health --endpoints=https://10.0.0.7:2379,https://10.0.0.9:2379,https://10.0.0.10:2379
	//	 --cacert=/etc/etcd/pki/ca.pem --cert=/etc/etcd/pki/etcd.cert --key=/etc/etcd/pki/etcd.key

}
