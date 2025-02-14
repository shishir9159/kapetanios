package etcd

import (
	"sync"
)

type ETCDClient struct {
	// TODO: check if this lock is necessary
	Mu sync.Mutex `json:"mu"`
}

func NewClient() *ETCDClient {

	return &ETCDClient{
		Mu: sync.Mutex{},
	}
}

func (client *ETCDClient) Healthcheck() error {

	return nil
}

func (client *ETCDClient) AddClient() {

}

func (client *ETCDClient) RemoveClient() {

}

func (client *ETCDClient) Cancel() {

}
