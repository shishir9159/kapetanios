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
	"strconv"
	"syscall"
	"time"
)

// TODO: REFACTOR REFACTOR REFACTOR

// kubelet error scraping

func getK8sCertsDir() string {

	// read from the configMap environment variable

	return "/etc/kubernetes/pki"
}

func getK8sConfigsDir() string {

	// TODO:
	//  read from the configmap populated by the
	//  lighthouse manager

	return "/etc/kubernetes/"
}

func getK8sConfigFiles() []string {
	// TODO:
	//  read from the configmap populated
	//  by lighthouse manager
	return []string{

		"admin.conf",
		"controller-manager.conf",
		"kubelet.conf",
		"scheduler.conf",
	}
}

func renameBackupDirectories(glob []string) error {

	for _, dir := range glob {
		index, err := strconv.Atoi(dir[11:])
		if err != nil {
			return err
		}

		err = os.Rename(dir, dir[11:]+strconv.Itoa(index+1))
		if err != nil {
			return err
		}
	}

	return nil
}

// checkList:
//    checks to ensure that access permissions works, directories exist
//    Check if copied files are the same using os.SameFile, return success if they are the same
//    Copy the bytes (all efficient means failed), return result

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherwise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.

func getLatestBackupDir() (string, error) {

	baseDir := "/opt/klovercloud/"
	latestBackup := "k8s-backup-1"
	backupDirectory := filepath.Join(baseDir, latestBackup)

	if baseDirFileInfo, err := os.Stat(baseDir); err != nil {
		if os.IsNotExist(err) {
			log.Println(baseDir, "backup directory for certificates doesn't exist")
			return "", fmt.Errorf("backup directory for certificate doesn't exist")
		} else if !baseDirFileInfo.IsDir() {
			// todo: remove the file and create a directory
		}
	} else if backupDirFileInfo, er := os.Stat(backupDirectory); er != nil {
		if os.IsNotExist(er) {
			log.Println(baseDir, "backups for certificates and kubeConfigs don't exist")
			return "", fmt.Errorf("no backup found")
		} else if !backupDirFileInfo.IsDir() {
			// todo: remove the file and create a directory
		}
	}

	return backupDirectory, nil
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

// It would not be optimal to get the list of files
// from the CopyDirectory, so a separate function parse
// through them again. Small repetition wins over complexity.
func fileChecklistValidation(backupDir string) []string {

	// todo:
	//  fileList

	return []string{""}
}

func overRideValidation(lastModifiedBeforeRollback time.Time) {
	stat, err := os.Stat("/etc/kubernetes/pki")
	if err != nil {
		fmt.Println("/etc/kubernetes/pki doesn't exist")
		return
	}

	if stat.ModTime() == lastModifiedBeforeRollback {
		fmt.Printf("last modification time difference %d", stat.ModTime().Unix()-lastModifiedBeforeRollback.Unix())
	}
}

func rollback(c Controller, connection pb.RollbackClient) error {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Info("Failed to create chroot on /host")
		return err
	}

	certsDir := getK8sCertsDir()
	kubeConfigs := getK8sConfigFiles()
	k8sConfigsDir := getK8sConfigsDir()
	backupDir, err := getLatestBackupDir()

	if err != nil {
		return err
	}

	stat, err := os.Stat("/etc/kubernetes/pki")
	if err != nil {
		return err
	}

	for _, kubeConfigFile := range kubeConfigs {
		// TODO:
		//  make sure override works
		//  or should I rename new and old
		srcFile := filepath.Join(backupDir, kubeConfigFile)
		destFile := filepath.Join(k8sConfigsDir, kubeConfigFile)
		er := Copy(srcFile, destFile)
		if er != nil {
			return er
		}
	}

	// TODO:
	//  make sure override works
	//  or should I rename new and old
	er := CopyDirectory(backupDir+"/pki/", certsDir)
	if er != nil {
		return er
	}

	overRideValidation(stat.ModTime())

	if err = changedRoot(); err != nil {
		c.log.Info("Failed to exit from the updated root")
		return err
	}

	rpc, err := connection.RollbackUpdate(c.ctx,
		&pb.RollbackStatus{
			PrerequisitesCheckSuccess: false,
			RollbackSuccess:           false,
			RestartSuccess:            false,
			RetryAttempt:              0,
			Log:                       "",
			Err:                       "",
		})

	if err != nil {
		c.log.Error("could not send status update: ",
			zap.Error(err))
	}

	c.log.Info("Status Update",
		zap.Bool("next step", rpc.GetProceedNextStep()),
		zap.Bool("terminate application", rpc.GetTerminateApplication()))

	return nil
}
