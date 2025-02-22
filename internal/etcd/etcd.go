package etcd

import (
	"flag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
)

type ETCD struct {
	// for external etcd nodes
	External struct {
		Endpoints []string `yaml:"endpoints"`
		CAFile    string   `yaml:"caFile"`
		CertFile  string   `yaml:"certFile"`
		KeyFile   string   `yaml:"keyFile"`
	} `yaml:"external"`
}

type ETCDClient struct {
	ETCD ETCD `yaml:"etcd"`
	conn net.Conn
}

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

func NewClient() *ETCDClient {
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(retryPolicy))
	if err != nil {
		c.log.Error().Err(err).
			Msg("couldn't connect to the kapetanios")

	}

	return &ETCDClient{}
}

func (client *ETCDClient) Healthcheck() error {

	return nil
}

func (client *ETCDClient) Cancel() {

}
