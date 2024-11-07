package main

import (
	pb "github.com/shishir9159/kapetanios/proto"
	"go.uber.org/zap"
)

func prerequisites(c Controller, connection pb.RollbackClient) error {

	// Contact the server and print out its response.

	rpc, err := connection.Prerequisites(c.ctx,
		&pb.PrerequisitesRollback{
			BackupExists:      true,
			SpaceAvailability: true,
			Err:               "",
		})

	if err != nil {
		c.log.Error("could not send status update: ", zap.Error(err))
	}

	c.log.Info("Status Update",
		zap.Bool("next step", rpc.GetProceedNextStep()),
		zap.Bool("terminate application", rpc.GetTerminateApplication()))

	return nil
}
