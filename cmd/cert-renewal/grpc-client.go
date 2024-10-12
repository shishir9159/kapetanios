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
	"io"
	"net"
	"net/http"
	"time"
)

const (
	defaultName = "cert-renewal"
)

func httpClient() {

	client := &http.Client{
		Timeout: time.Second * 5,
		Transport: &http.Transport{
			// Avoid: "x509: certificate signed by unknown authority"
			//TLSClientConfig: &tls.Config{
			//	InsecureSkipVerify: true,
			//},
			// Inspect the network connection type
			DialContext: func(ctx context.Context, network string, addr string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "tcp4", addr)
			},

			//DialContext: (&net.Dialer{
			//	Control: func(network, address string, c syscall.RawConn) error {
			//		// Reference: https://golang.org/pkg/net/#Dial
			//		if network == "tcp4" {
			//			log.Error("ipv6 error")
			//		}
			//		return nil
			//	},
			//}).DialContext,
		},
	}

	req, err := http.NewRequest("GET", "http://hello.default.svc.cluster.local", nil)
	if err != nil {
		log.Error("Failed to connect to hello.default.svc.cluster.local", zap.Error(err))
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Error("error from client request", zap.Error(err))
	}

	if resp == nil {
		log.Fatal("response is empty")
	}

	defer func(Body io.ReadCloser) {
		er := Body.Close()
		if er != nil {
			log.Error("error closing the body")
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failed to read body", zap.Error(err))
	}

	if body != nil {
		log.Info("body", zap.String("body", string(body)))
	}

}

var (
	//addr = flag.String("addr", "kapetanios-grpc.com:80", "the address to connect to") .svc.cluster.local
	//addr = flag.String("addr", "dns:[//10.96.0.1/]kapetanios.default.svc.cluster.local[:50051]", "the address to connect to")
	name = flag.String("name", defaultName, "gRPC test")
)

func GrpcClient(log *zap.Logger) {

	var addr = flag.String("addr", "kapetanios.default.svc.cluster.local:50051", "the address to connect to")

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
		log.Info("address", zap.String("addr", *addr))
	}

	if err != nil {
		log.Error("error kapetanios address", zap.Error(err))
	}

	flag.Parse()

	log.Info("connecting to kapetanios service address", zap.String("addr", *addr))
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

	c := pb.NewRenewalClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	rpc, err := c.StatusUpdate(ctx, &pb.CreateRequest{BackupSuccess: true, RenewalSuccess: true, RestartSuccess: true, Log: "", Err: "s"})
	if err != nil {
		log.Error("could not send status update: ", zap.Error(err))
	}

	log.Info("Status Update", zap.Bool("next step", rpc.GetProceedNextStep()), zap.Bool("retry", rpc.GetRetryCurrentStep()))
}
