package main

import (
	"log"
)

//const (
//	addr = flag.String("addr", "localhost:50051", "the address to connect to")
//	name = flag.String("name", "certs", "Name to greet")
//)

func main() {

	//flag.Parse()
	//// Set up a connection to the server.
	//conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	//if err != nil {
	//	log.Fatalf("did not connect: %v", err)
	//}
	//
	//defer func(conn *grpc.ClientConn) {
	//	er := conn.Close()
	//	if er != nil {
	//
	//	}
	//}(conn)
	//c := pb.NewGreeterClient(conn)
	//
	//// Contact the server and print out its response.
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	//r, err := c.SayHello(ctx, &pb.HelloRequest{Name: *name})
	//if err != nil {
	//	log.Fatalf("could not greet: %v", err)
	//}
	//log.Printf("Greeting: %s", r.GetMessage())

	//	step 1. Backup directories
	err := BackupCertificatesKubeConfigs(7)
	if err != nil {
		log.Println(err)
	}

	//	step 2.
	err = Renew()
	if err != nil {
		log.Println(err)
	}

	//step 3.
	err = Restart()
	if err != nil {
		log.Println(err)
	}
}
