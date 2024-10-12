package main

import (
	"context"
	"flag"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net"
	"net/http"
	"time"
)

const (
	defaultName = "cert-renewal"
)

//type Post struct {
//	"message" string `json:"message"`
//}

var (
	//addr = flag.String("addr", "kapetanios-grpc.com:80", "the address to connect to") .svc.cluster.local
	addr = flag.String("addr", "kapetanios.default.svc.cluster.local:50051", "the address to connect to")
	//addr = flag.String("addr", "dns:[//10.96.0.1/]kapetanios.default.svc.cluster.local[:50051]", "the address to connect to")
	name = flag.String("name", defaultName, "gRPC test")
)

func GrpcClient(log *zap.Logger) {

	flag.Parse()
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

	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, network, "10.96.0.10:53")
		},
	}

	ip, err := r.LookupIP(context.Background(), "ip4", "www.google.com")
	if len(ip) != 0 {
		log.Info("google address")
		fmt.Println(ip)
	}

	if err != nil {
		log.Error("error google address", zap.Error(err))
		fmt.Println(ip)
	}

	ip, err = r.LookupIP(context.Background(), "ip4", "http://hello.default.svc.cluster.local")
	if len(ip) != 0 {
		log.Info("hello http address")
		fmt.Println(ip)
	}

	if err != nil {
		log.Error("error hello http address", zap.Error(err))
		fmt.Println(ip)
	}

	ip, err = r.LookupIP(context.Background(), "ip4", "hello.default.svc.cluster.local")
	if len(ip) != 0 {
		log.Info("hello service address")
		fmt.Println(ip)
	}

	if err != nil {
		log.Error("error hello http address")
	}

	ip, err = r.LookupIP(context.Background(), "ip4", "kapetanios.default.svc.cluster.local")
	if len(ip) != 0 {
		log.Info("hello service address")
		fmt.Println(ip)
	}

	if err != nil {
		log.Error("error kapetanios address")
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

	log.Info("body", zap.String("body", string(body)))

	//address, err := net.LookupHost("kapetanios.default.svc.cluster.local")
	//
	//if err != nil {
	//	log.Error("failed to resolve host", zap.String("address", address[0]), zap.String("address", address[1]), zap.Error(err))
	//}
	//
	//if address != nil {
	//	log.Info("resolved host", zap.String("host", address[0]), zap.String("host", *addr))
	//	addr = flag.String("addr", address[0]+"50051", "the address to connect to")
	//	addr = &address[0]
	//}
	//
	//// Set up a connection to the server.
	//conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	//if err != nil {
	//	log.Error("did not connect", zap.Error(err))
	//}
	////grpc.WithDisableServiceConfig()
	//defer func(conn *grpc.ClientConn) {
	//	er := conn.Close()
	//	if er != nil {
	//		log.Error("failed to close the grpc connection",
	//			zap.Error(er))
	//	}
	//}(conn)
	//
	//c := pb.NewRenewalClient(conn)
	//
	//// Contact the server and print out its response.
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	//r, err := c.StatusUpdate(ctx, &pb.CreateRequest{BackupSuccess: true, RenewalSuccess: true, RestartSuccess: true, Log: "", Err: "s"})
	//if err != nil {
	//	log.Error("could not send status update: ", zap.Error(err))
	//}
	//log.Error("Status Update", zap.Bool("next step", r.GetProceedNextStep()), zap.Bool("retry", r.GetRetryCurrentStep()))
}
