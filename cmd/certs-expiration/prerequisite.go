package main

import (
	"fmt"
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"go.uber.org/zap"
	"io"
	"os"
	"syscall"
)

// todo:
//  move to internal/ cluster health check

// todo: check if localAPIEndpoint is not same across the nodes

// todo: refactor with interface

// equivalent to `df /path/to/mounted/filesystem`
// it works with filesystems not device files
// how much data could be stored in the filesystem mounted at provided path
// warning -- filesystems have overheads
func getStorageStat(path string) (int64, error) {

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

// equivalent to `lsblk --bytes /dev/device`
// to calculate the total storable data amount in th block device
func getStorage(path string) (int64, error) {

	// location exists or not?
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("error opening %s: %s\n", path, err)
		os.Exit(1)
	}

	// meant to work with block devices todo: trial and error
	pos, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		fmt.Printf("error seeking to end of %s: %s\n", path, err)
		os.Exit(1)
	}

	return pos / 1048576, nil
}

func NodeHealth(c Controller, connection pb.ValidityClient) error {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Fatal("Failed to create chroot on /host",
			zap.Error(err))
		return err
	}

	freeSpace, err := getStorage("/opt/")

	if err != nil {
		c.log.Error("Failed to get storage space for /opt/ directory", zap.Error(err))
		return err
	}

	// TODO:
	if freeSpace != 0 {
		c.log.Info("available free space in the /opt/ directory", zap.Int64("freeSpace in MB ", freeSpace/1048576))
	}

	// TODO:
	//  read where is data directory for etcd
	//  Data Dir is
	freeSpace, err = getStorage("/var/lib/")

	if err != nil {
		c.log.Error("Failed to get storage space for /var/lib/etcd directory", zap.Error(err))
		return err
	}

	// TODO:
	if freeSpace != 0 {
		c.log.Info("available free space in the /var/lib/etcd directory", zap.Int64("freeSpace in MB ", freeSpace/1048576))
	}

	//	etcd status
	//

	if err = changedRoot(); err != nil {
		c.log.Fatal("Failed to exit from the updated root",
			zap.Error(err))

		return err
	}

	rpc, err := connection.ClusterHealthChecking(c.ctx,
		&pb.PrerequisitesExpiration{
			EtcdStatus:   true,
			DiskPressure: false,
			Err:          "",
		})

	if err != nil {
		c.log.Error("could not send status update: ", zap.Error(err))
	}

	c.log.Info("Status Update",
		zap.Bool("response received", rpc.GetResponseReceived()))

	return nil
}
