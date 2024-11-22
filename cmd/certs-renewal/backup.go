package main

import (
	"fmt"
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"go.uber.org/zap"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"syscall"
)

func getK8sConfigsDir() string {

	// TODO:
	//  read from the configmap populated
	//  by lighthouse manager
	return "/etc/kubernetes/"
}

func renameBackupDirectories(s int, glob []string) error {

	for _, dir := range glob {
		log.Println("dir:" + dir)
		log.Println("dir[s:]:" + dir[s:])
		index, err := strconv.Atoi(dir[s:])
		if err != nil {
			return err
		}

		err = os.Rename(dir, dir[s:]+strconv.Itoa(index+1))
		if err != nil {
			return err
		}
	}

	return nil
}

func checkSurplusBackupDirs(backupCount int, baseDir string, backupDirPattern string) (int, error) {

	s := len(baseDir) + len(backupDirPattern) - 1

	glob, err := filepath.Glob(baseDir + backupDirPattern + "*")
	if err != nil {
		log.Println(glob)
		log.Println(err)
	}

	fmt.Println("glob 1:")
	fmt.Println(glob)

	if len(glob) == 0 {
		return 1, nil
	}
	// natural sorting assumes the
	// backupDirPattern is of s letters
	sort.Slice(glob, func(i, j int) bool {
		if glob[i][:s] != glob[j][:s] {
			return glob[i] < glob[j]
		}
		ii, _ := strconv.Atoi(glob[i][s:])
		jj, _ := strconv.Atoi(glob[j][s:])
		return ii < jj
	})

	fmt.Println("glob 2:")
	fmt.Println(glob)

	if len(glob) >= backupCount {
		er := removeDirectory(glob[backupCount-1])
		if er != nil {
			return 1, er
		}
	}

	err = renameBackupDirectories(s, glob)
	if err != nil {
		return 1, err
	}

	return 1, nil
}

// checkList:
//    checks to ensure that access permissions works, directories exist
//    Check if copied files are the same using os.SameFile, return success if they are the same
//    Copy the bytes (all efficient means failed), return result

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherwise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.

func getBackupDir(backupCount int) (string, error) {

	baseDir := "/opt/klovercloud/"
	backupDirPattern := "k8s-backup-"
	//certsBackupDirPattern := "certs-backup-"

	if dfi, err := os.Stat(baseDir); err != nil {
		if os.IsNotExist(err) {
			log.Println(baseDir, "creating backup directory for certificates")
			if er := CreateIfNotExists(baseDir, 0755); er != nil {
				log.Println(er)
				return "", er
			}
		} else if !dfi.IsDir() {
			// remove the file and create a directory
		}
	}

	// TO-DO: handle possible permission errors for file copy
	//if err != nil {
	//	return ""
	//}

	//	if sfi, err := os.Stat(baseDir); os.IsNotExist(err) {
	//		fmt.Println(baseDir, "creating backup directory for certificates")
	//		if err := CreateIfNotExists(baseDir, 0755); err != nil {
	//			return "", err
	//		}
	//	}

	index, err := checkSurplusBackupDirs(backupCount, baseDir, backupDirPattern)
	if err != nil {
		fmt.Println(err)
	}

	// permission 600 or 0755
	if er := CreateIfNotExists(baseDir+backupDirPattern+strconv.Itoa(index), 0755); er != nil {
		log.Println(er)
		return "", er
	}

	return baseDir + backupDirPattern + strconv.Itoa(index), nil
}

func CopyDirectory(src, dest string) error {

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {

		sourcePath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, er := os.Stat(sourcePath)
		if er != nil {
			return er
		}

		stat, ok := fileInfo.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if e := CreateIfNotExists(destPath, 0755); e != nil {
				return e
			}
			if e := CopyDirectory(sourcePath, destPath); e != nil {
				return e
			}
		case os.ModeSymlink:
			if e := CopySymLink(sourcePath, destPath); e != nil {
				return e
			}
		//	sfi.Mode().IsRegular() with os.Stat() can help determine if the srcDir is a device and should be avoided
		default:
			if e := Copy(sourcePath, destPath); e != nil {
				return e
			}
		}

		if e := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); e != nil {
			return e
		}

		fInfo, er := entry.Info()
		if er != nil {
			return er
		}

		isSymlink := fInfo.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if e := os.Chmod(destPath, fInfo.Mode()); e != nil {
				return e
			}
		}
	}

	return nil
}

func Copy(srcFile, dstFile string) error {

	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}

	defer func(out *os.File) {
		er := out.Close()
		if er != nil {

		}
	}(out)

	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}

	defer func(in *os.File) {
		er := in.Close()
		if er != nil {

		}
	}(in)

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

func removeDirectory(dirPath string) error {

	err := os.RemoveAll(dirPath)
	if err != nil {
		return fmt.Errorf("failed to remove directory recursively: %w", err)
	}

	return nil
}

func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

func CreateIfNotExists(dir string, perm os.FileMode) error {
	if Exists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}

func CopySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return err
	}
	return os.Symlink(link, dest)
}

func etcdBackup() {
	//	ETCDCTL_API=3 etcdctl --endpoints=https://127.0.0.1:2379 \
	//--cacert=/etc/kubernetes/pki/etcd/ca.crt \
	//--cert=/etc/kubernetes/pki/etcd/server.crt \
	//--key=/etc/kubernetes/pki/etcd/server.key \
	//snapshot save <backup-file-location>

}

// It would not be optimal to get the list of files
// from the CopyDirectory, so a separate function parse
// through them again. Small repetition wins over complexity.
func fileChecklistValidation(backupDir string) []string {

	// todo:
	//  fileList

	return []string{""}
}

// todo:
//  compression enabled copy

func BackupCertificatesKubeConfigs(c Controller, backupCount int, connection pb.RenewalClient) (bool, error) {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Error("Failed to create chroot on /host",
			zap.Error(err))
		return false, err
	}

	backupDir, err := getBackupDir(backupCount)
	kubernetesConfigDir := getK8sConfigsDir()

	// Copy Recursively
	err = CopyDirectory(kubernetesConfigDir, backupDir)
	if err != nil {
		c.log.Error("Failed to take backups",
			zap.Error(err))
		// todo: send the error the server: return false, err
	}

	failedCopyFiles := fileChecklistValidation(backupDir)
	if len(failedCopyFiles) != 0 {
		c.log.Error("Failed to take backups",
			zap.Error(err))
		// todo: send the error the server: return false, err
	}

	if err = changedRoot(); err != nil {
		c.log.Fatal("Failed to exit from the updated root",
			zap.Error(err))
	}

	rpc, err := connection.BackupUpdate(c.ctx,
		&pb.BackupStatus{
			EtcdBackupSuccess:       true,
			KubeConfigBackupSuccess: true,
			FileChecklistValidation: true,
			Err:                     "",
		})

	if err != nil {
		c.log.Error("could not send status update: ", zap.Error(err))
	}

	c.log.Info("Backup Status",
		zap.Bool("next step", rpc.GetProceedNextStep()),
		zap.Bool("terminate application", rpc.GetTerminateApplication()))

	return rpc.GetProceedNextStep(), err
}
