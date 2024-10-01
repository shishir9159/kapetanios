package main

import (
	"fmt"
	"time"
)

func run(ch chan struct{}) {
	fmt.Println("run")
	time.Sleep(1 * time.Second)
	ch <- struct{}{}
}

func RunForever() {
	wait := make(chan struct{})
	for {
		go run(wait)
		<-wait
	}
}

func main() {

	Cert("default")
	RunForever()

	//client, err := orchestration.NewClient()
	//if err != nil {
	//	fmt.Println("Error creating Kubernetes client: %v", err)
	//}
	//
	//renewalAgentManager := orchestration.NewAgent(client)
	//
	//nodeRole := "etcd"
	//pod, err := renewalAgentManager.CreateTempPod(context.Background(), nodeRole)
	//if err != nil {
	//	fmt.Println("Error creating temporary pod: %v", err)
	//}

	//	what happens when lighthouse fails in the middle of the cert renewal process?

}
