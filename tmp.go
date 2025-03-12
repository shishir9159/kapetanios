package main

import (
	"context"
	"encoding/json"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
)

type Receiver interface {
	reader()
}

type Sender struct {
	SomeOtherValue string `json:"someOtherValue"`
}

func (s Sender) reader() {
	fmt.Println("quick pretend to read something")
	fmt.Println(s.SomeOtherValue)
}

type Messenger struct {
	SomeValue string   `json:"someValue"`
	Receiver  Receiver `json:"receiver"`
}

func main() {

	kubeConfigPath := clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
	config, err := clientcmd.BuildConfigFromFlags("", "./config")
	if err != nil {
		log.Print(fmt.Errorf(
			"unable to load kubeconfig from %s: %v",
			kubeConfigPath, err))
	}

	clientSet, err := kubernetes.NewForConfig(config)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if clientSet != nil {

		configMap, err := clientSet.CoreV1().ConfigMaps("default").
			Get(ctx, "kapetanios", metav1.GetOptions{})

		if err != nil {
			log.Print(err)
		}

		s := Sender{
			SomeOtherValue: "something in the way",
		}

		messenger := Messenger{
			SomeValue: "light will guide you home",
			Receiver:  s,
		}

		payload, err := json.Marshal(&messenger)
		configMap.Data["kapetanios"] = string(payload)

		_, err = clientSet.CoreV1().ConfigMaps("default").
			Update(ctx, configMap, metav1.UpdateOptions{})

		m := Messenger{}

		load := []byte(configMap.Data["kapetanios"])
		err = json.Unmarshal(load, &m)
		if err != nil {
			log.Print(m.Receiver)
			log.Print("---")
			log.Print(string(load))
			log.Print("---")
			log.Print(load)
		}
	} else {
		log.Println("clientSet is nil")
	}
}
