package main

import (
	"go.uber.org/zap"
	"log"
	"syscall"
)

// check if localAPIEndpoint is not same across the nodes

// todo:
//	how to leverage lseek to get block storage space
//	https://stackoverflow.com/questions/46558824/how-do-i-get-a-block-devices-size-correctly-in-go

func getStorage(path string) (int64, error) {

	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)

	log.Printf("Total Disk Space: %d", stat.Blocks*uint64(stat.Bsize))
	log.Printf("Avail Disk Space: %d", stat.Bavail*uint64(stat.Bsize))
	log.Printf("Free Disk Space: %d", stat.Bfree*uint64(stat.Bsize))
	log.Printf("Used Disk Space: %d", stat.Blocks*uint64(stat.Bsize)-stat.Bfree*uint64(stat.Bsize))

	if err != nil {
		return 0, err
	}

	return int64(stat.Bfree) * stat.Bsize, nil
}

func PrerequisitesForCertRenewal(log *zap.Logger) error {

	// available disk space
	freeSpace, err := getStorage("/opt/")

	if err != nil {
		log.Error("Failed to get storage space for /opt/ directory", zap.Error(err))
		return err
	}

	// TODO:
	if freeSpace != 0 {
		log.Info("available free space in the /opt/ directory", zap.Int64("freeSpace in MB: ", freeSpace/1048576))
	}

	// TODO:
	//  read where is data directory for etcd
	//  Data Dir is
	freeSpace, err = getStorage("/var/lib/")

	if err != nil {
		log.Error("Failed to get storage space for /var/lib/etcd directory", zap.Error(err))
		return err
	}

	// TODO:
	if freeSpace != 0 {
		log.Info("available free space in the /var/lib/etcd directory", zap.Int64("freeSpace in MB: ", freeSpace/1048576))
	}

	//	etcd status
	//

	return nil
}
