package main

import (
	"flag"
	"go.uber.org/zap"
	"io"
	"net/http"
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

	resp, err := http.Get("hello.default.svc.cluster.local")
	if err != nil {
		log.Error("Failed to connect to hello.default.svc.cluster.local", zap.Error(err))
	}
	defer resp.Body.Close()
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
