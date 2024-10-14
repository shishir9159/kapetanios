package main

import (
	"go.uber.org/zap"
	"syscall"
)

// check if localAPIEndpoint is not same across the nodes
func getStorage(path string) (int64, error) {

	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return 0, err
	}

	return int64(stat.Bfree) * stat.Bsize, nil
}

func PrerequisitesForCertRenewal(log *zap.Logger) {

	// available disk space
	freeSpace, err := getStorage("/opt/")

	if err != nil {
		log.Error("Failed to get storage space for /opt/ directory", zap.Error(err))
	}

	// TODO:
	if freeSpace != 0 {
		log.Info("Free space", zap.Int64("freeSpace", freeSpace))
	}

	// TODO:
	//  read where is data directory for etcd
	//  Data Dir is
	freeSpace, err = getStorage("/var/lib/etcd")

	if err != nil {
		log.Error("Failed to get storage space for /opt/ directory", zap.Error(err))
	}

	// TODO:
	if freeSpace != 0 {
		log.Info("Free space", zap.Int64("freeSpace", freeSpace))
	}

	//	etcd status
	//
}
