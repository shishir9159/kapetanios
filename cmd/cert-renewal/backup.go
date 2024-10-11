package main

import (
	"context"
	"fmt"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"github.com/shishir9159/kapetanios/utils"
	"go.uber.org/zap"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

func getK8sConfigsDir() string {
	//	will be called after getKubeadmFileLocation
	// default:	/etc/kubernetes/pki

	// etcd certs and nodes info
	//etcd:
	//  external:
	//    caFile: /etc/kubernetes/pki/etcd-ca.pem
	//    certFile: /etc/kubernetes/pki/etcd.cert
	//    endpoints:
	//    - https://5.161.64.103:2379
	//    - https://5.161.248.112:2379
	//    - https://5.161.67.249:2379
	//    keyFile: /etc/kubernetes/pki/etcd.key
	//kubernetesVersion

	// format Data:map[string]string{ClusterConfiguration

	client, err := orchestration.NewClient()
	if err != nil {
		fmt.Printf("Error creating Kubernetes client: %v\n", err)
		return "/etc/kubernetes/"
		//return "/etc/kubernetes/pki"
	}

	// todo: should be relocated to lighthouse agent
	cm, err := client.Clientset().CoreV1().ConfigMaps("kube-system").Get(context.Background(), "kubeadm-config", metav1.GetOptions{})

	if err != nil {
		fmt.Printf("Error getting kubeadm-config: %v\n", cm)
	}

	return "/etc/kubernetes/"

	// convert the cm to a file and read from the yaml file
	//var certsDir string = cm.BinaryData // certificatesDir

}

func checkSurplusBackupDirs(backupCount int, baseDir string, backupDirPattern string) (int, error) {

	glob, err := filepath.Glob(baseDir + backupDirPattern + "*")
	if err != nil {

		log.Println(err)
	}

	if len(glob) >= backupCount {
		// increment the indices
		//sort.Slice(s, func(i, j int) bool {
		//		if s[i][:2] != s[j][:2] {
		//			return s[i] < s[j]
		//		}
		//		ii, _ := strconv.Atoi(s[i][2:])
		//		jj, _ := strconv.Atoi(s[j][2:])
		//		return ii < jj
		//	})
	}

	return len(glob) + 1, nil
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

}

// todo:
//  compression enabled copy

func BackupCertificatesKubeConfigs(c Controller, backupCount int) error {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Error("Failed to create chroot on /host",
			zap.Error(err))
		return err
	}

	backupDir, err := getBackupDir(backupCount)
	kubernetesConfigDir := getK8sConfigsDir()

	// Copy Recursively
	err = CopyDirectory(kubernetesConfigDir, backupDir)
	if err != nil {
		c.log.Error("Failed to take backups",
			zap.Error(err))
	}

	if err = changedRoot(); err != nil {
		c.log.Fatal("Failed to exit from the updated root",
			zap.Error(err))
	}

	return err
}
