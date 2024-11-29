package main

import (
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"syscall"
)

// todo:
//  move to internal/ cluster health check
//  check if localAPIEndpoint is not same across the nodes

func etcdStatus() ([]string, error) {

	return nil, nil
}

// todo:
//	 how to leverage lseek to get block storage space
//	 https://stackoverflow.com/questions/46558824/how-do-i-get-a-block-devices-size-correctly-in-go

func getStorage(path string) (int64, error) {

	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)

	//log.Printf("Total Disk Space: %d", stat.Blocks*uint64(stat.Bsize)/1048576)
	//log.Printf("Avail Disk Space: %d", stat.Bavail*uint64(stat.Bsize/1048576))
	//log.Printf("Free Disk Space: %d", stat.Bfree*uint64(stat.Bsize)/1048576)
	//log.Printf("Used Disk Space: %d", stat.Blocks*uint64(stat.Bsize)/1048576-stat.Bfree*uint64(stat.Bsize)/1048576)

	if err != nil {
		return 0, err
	}

	return int64(stat.Bfree) * stat.Bsize, nil
}

func PrerequisitesForCertRenewal(c Controller, connection pb.RenewalClient) (bool, error) {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to create chroot on /host")
	}

	freeSpace, err := getStorage("/opt/")

	if err != nil {
		c.log.Error().Err(err).
			Msg("failed to get storage space for /opt/ directory")

		// TODO: fetch response to proceed or not
		//  return false, err
		// TODO: concatenate error string
	}

	// TODO:
	if freeSpace != 0 {
		c.log.Info().
			Int64("free space(MB)", freeSpace/1048576).
			Msg("available free space in the /opt/ directory")
	}

	// TODO:
	//  read where is data directory for etcd
	//  Data Dir is
	freeSpace, err = getStorage("/var/lib/")

	if err != nil {
		c.log.Error().Err(err).
			Msg("failed to get storage space for /var/lib/ directory")

		// TODO: fetch response to proceed or not
		//  return false, err
		//	TODO: concatenate error string
	}

	// TODO:
	if freeSpace != 0 {
		c.log.Info().
			Int64("free space(MB)", freeSpace/1048576).
			Msg("available free space in the /var/lib/ directory")
	}

	// TODO: etcd status
	//  sudo ETCDCTL_API=3 etcdctl endpoint status --endpoints=https://10.0.0.7:2379,https://10.0.0.9:2379,https://10.0.0.10:2379
	//  --cacert=/etc/etcd/pki/ca.pem --cert=/etc/etcd/pki/etcd.cert --key=/etc/etcd/pki/etcd.key

	if err = changedRoot(); err != nil {
		c.log.Fatal().Err(err).
			Msg("Failed to exit from the updated root")
	}

	rpc, err := connection.ClusterHealthChecking(c.ctx,
		&pb.PrerequisitesRenewal{
			EtcdStatus:             true,
			ExternallyManagedCerts: false, //TODO - process the output
			EtcdDirFreeSpace:       50,    //TODO: realistic value
			KubeDirFreeSpace:       50,
			LocalAPIEndpoint:       "",
			Err:                    "",
		})

	if err != nil {
		c.log.Error().Err(err).
			Msg("could not send status update")
	}

	c.log.Info().
		Bool("next step", rpc.GetProceedNextStep()).
		Bool("terminate application", rpc.GetTerminateApplication()).
		Msg("status update")

	return rpc.GetProceedNextStep(), nil
}
