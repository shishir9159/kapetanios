package main

import (
	"fmt"
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
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
		// TODO: use controller
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
		c.log.Fatal().Err(err).
			Msg("Failed to create chroot on /host")
	}

	freeSpace, err := getStorage("/opt/")
	if err != nil {
		c.log.Error().Err(err).
			Msg("failed to get storage space for /opt/ directory")
		return err
	}

	// TODO:
	if freeSpace != 0 {
		// TODO: fast formatting to error
		c.log.Error().Err(nil).
			Int64("available free space in /opt directory", freeSpace).
			Msg("not enough free space for a healthy cluster")
	}

	// TODO:
	//  read where is data directory for etcd
	//  Data Dir is
	freeSpace, err = getStorage("/var/lib/")

	if err != nil {
		c.log.Error().Err(err).
			Msg("failed to get storage space for /var/lib/ directory")
		// todo:
		//  return err
	}

	// TODO:
	if freeSpace != 0 {
		c.log.Error().Err(nil).
			Int64("available free space in /var/lib directory", freeSpace).
			Msg("not enough free space for a healthy cluster")
	}

	//	etcd status
	//

	if err = changedRoot(); err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to exit from the updated root")
	}

	rpc, err := connection.ClusterHealthChecking(c.ctx,
		&pb.PrerequisitesExpiration{
			EtcdStatus:   true,
			DiskPressure: false,
			Err:          "",
		})

	if err != nil {
		c.log.Error().Err(err).
			Msg("could not send status update")
	}

	c.log.Info().
		Bool("response received", rpc.GetResponseReceived()).
		Msg("status update")

	return nil
}
