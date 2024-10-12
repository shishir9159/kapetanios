package main

//
//import (
//	"context"
//	"flag"
//	"fmt"
//	"google.golang.org/grpc"
//	"google.golang.org/grpc/codes"
//	pb "google.golang.org/grpc/examples/features/proto/echo"
//	"google.golang.org/grpc/status"
//	"log"
//	"net"
//	"sync"
//)
//
////var port = flag.Int("port", 50052, "port number")
//
//type failingServer struct {
//	pb.UnimplementedEchoServer
//	mu sync.Mutex
//
//	reqCounter uint
//	reqModulo  uint
//}
//
//// this method will fail reqModulo - 1 times RPCs and return status code Unavailable,
//// and succeeded RPC on reqModulo times.
//func (s *failingServer) maybeFailRequest() error {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//	s.reqCounter++
//	if (s.reqModulo > 0) && (s.reqCounter%s.reqModulo == 0) {
//		return nil
//	}
//
//	return status.Errorf(codes.Unavailable, "maybeFailRequest: failing it")
//}
//
//func (s *failingServer) UnaryEcho(_ context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
//	if err := s.maybeFailRequest(); err != nil {
//		log.Println("request failed count:", s.reqCounter)
//		return nil, err
//	}
//
//	log.Println("request succeeded count:", s.reqCounter)
//	return &pb.EchoResponse{Message: req.Message}, nil
//}
//
//func grpcCert() {
//
//	flag.Parse()
//
//	address := fmt.Sprintf(":%v", *port)
//	lis, err := net.Listen("tcp", address)
//	if err != nil {
//		log.Fatalf("failed to listen: %v", err)
//	}
//	fmt.Println("listen on address", address)
//
//	s := grpc.NewServer()
//
//	// Configure server to pass every fourth RPC;
//	// client is configured to make four attempts.
//	failingservice := &failingServer{
//		reqCounter: 0,
//		reqModulo:  4,
//	}
//
//	pb.RegisterEchoServer(s, failingservice)
//	if err := s.Serve(lis); err != nil {
//		log.Fatalf("failed to serve: %v", err)
//	}
//
//}
