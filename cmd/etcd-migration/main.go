package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
)

type Controller struct {
	ctx context.Context
	log *zap.Logger
}

func main() {

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Println(err)
	}

	//zap.ReplaceGlobals(logger)

	// replace zap with zerolog

	c := Controller{
		ctx: context.Background(),
		log: logger,
	}

	defer func(logger *zap.Logger) {
		er := logger.Sync()
		if er != nil {
			c.log.Error("failed to close logger",
				zap.Error(er))
		}
	}(logger)

	//	TODO: etcd remove
	//	 ETCDCTL_API=3 etcdctl endpoint health --endpoints=https://10.0.0.7:2379,https://10.0.0.9:2379,https://10.0.0.10:2379
	//	 --cacert=/etc/etcd/pki/ca.pem --cert=/etc/etcd/pki/etcd.cert --key=/etc/etcd/pki/etcd.key

}
