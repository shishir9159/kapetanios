package main

//import (
//	"context"
//	"google.golang.org/grpc"
//	"log"
//	"time"
//
//	"google.golang.org/grpc/credentials/insecure"
//	pb "google.golang.org/grpc/examples/features/proto/echo"
//)
//
//var (
//	//addr        = flag.String("addr", "localhost:50052", "the address to connect to")
//	retryPolicy = `{
//		"methodConfig": [{
//		  "name": [{"service": "grpc.examples.echo.Echo"}],
//		  "retryPolicy": {
//			  "MaxAttempts": 4,
//			  "InitialBackoff": ".01s",
//			  "MaxBackoff": ".01s",
//			  "BackoffMultiplier": 1.0,
//			  "RetryableStatusCodes": [ "UNAVAILABLE" ]
//		  }
//		}]}`
//)
//
//func grpcClient() {
//
//	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultServiceConfig(retryPolicy))
//	if err != nil {
//		log.Fatalf("did not connect: %v", err)
//	}
//	defer func() {
//		if e := conn.Close(); e != nil {
//			log.Printf("failed to close connection: %s", e)
//		}
//	}()
//
//	c := pb.NewEchoClient(conn)
//
//	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
//	defer cancel()
//
//	reply, err := c.UnaryEcho(ctx, &pb.EchoRequest{Message: "Try and Success"})
//	if err != nil {
//		log.Fatalf("UnaryEcho error: %v", err)
//	}
//	log.Printf("UnaryEcho reply: %v", reply)
//
//}
