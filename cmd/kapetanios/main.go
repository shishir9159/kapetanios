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

	//	what happens when lighthouse fails in the middle of the cert renewal process?

}
