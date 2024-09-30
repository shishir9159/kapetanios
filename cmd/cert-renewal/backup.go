package main

import (
	"context"
	"fmt"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

func getCertificatesDir() string {
	//	will be called after getKubeadmFileLocation
	// default:	/etc/kubernetes/pki

	client, err := orchestration.NewClient()
	if err != nil {
		fmt.Printf("Error creating Kubernetes client: %v\n", err)
		return "/etc/kubernetes/pki"
	}

	cm, err := client.Clientset().CoreV1().ConfigMaps("kube-system").Get(context.Background(), "kubeadm-config", metav1.GetOptions{})

	// temporary lines
	print(cm)
	return "/etc/kubernetes/pki"

	// convert the cm to a file and read from the yaml file
	//var certsDir string = cm.BinaryData // certificatesDir

}

func getKubeConfigFiles() []string {

	// could be extracted from the static manifest files
	// --kubeconfig=/etc/kubernetes/controller-manager.conf

	kubeConfigFiles := []string{
		"/etc/kubernetes/admin.conf",
		"/etc/kubernetes/controller-manager.conf",
		"/etc/kubernetes/scheduler.conf",
	}

	return kubeConfigFiles
}

// checkList:
//    checks to ensure that access permissions works, directories exist
//    Check if copied files are the same using os.SameFile, return success if they are the same
//    Copy the bytes (all efficient means failed), return result

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherwise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.

func getBackupDir(backupCount int) (string, error) {

	baseDir := "/opt/klovercloud"
	backupDirPattern := "certs-backup-*"

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

	// TO-DO: handle possible permission errors for file coppy
	//if err != nil {
	//	return ""
	//}

	//	if sfi, err := os.Stat(baseDir); os.IsNotExist(err) {
	//		fmt.Println(baseDir, "creating backup directory for certificates")
	//		if err := CreateIfNotExists(baseDir, 0755); err != nil {
	//			return "", err
	//		}
	//	}

	glob, err := filepath.Glob(baseDir + backupDirPattern)
	log.Println("glob", glob)
	if err != nil {

		log.Println(err)
	}

	if len(glob) == 0 {
		if er := CreateIfNotExists(baseDir+backupDirPattern+"1", 0755); er != nil {
			// permission 600 or 0755
			log.Println(er)
			return baseDir + backupDirPattern + "1", er
		}
	} else if len(glob) < backupCount {
		if er := CreateIfNotExists(baseDir, 0755); er != nil {
			log.Println(er)
			return baseDir + backupDirPattern + string(rune(len(glob)+1)), er
		}
		return baseDir + backupDirPattern + string(rune(len(glob)+1)), nil
	} else {
		//	logic for removing the oldest one and increment all indices by one
	}

	return "/opt/klovercloud/certs-backup-1", nil
}

func CopyDirectory(src, dest string) error {

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {

		sourcePath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		stat, ok := fileInfo.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := CreateIfNotExists(destPath, 0755); err != nil {
				return err
			}
			if err := CopyDirectory(sourcePath, destPath); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err := CopySymLink(sourcePath, destPath); err != nil {
				return err
			}
		//	sfi.Mode().IsRegular() with os.Stat() can help determine if the srcDir is a device and should be avoided
		default:
			if err := Copy(sourcePath, destPath); err != nil {
				return err
			}
		}

		if err := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); err != nil {
			return err
		}

		fInfo, err := entry.Info()
		if err != nil {
			return err
		}

		isSymlink := fInfo.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, fInfo.Mode()); err != nil {
				return err
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
		err := out.Close()
		if err != nil {

		}
	}(out)

	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}

	defer func(in *os.File) {
		err := in.Close()
		if err != nil {

		}
	}(in)

	_, err = io.Copy(out, in)
	if err != nil {
		return err
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

func BackupCertificatesKubeConfigs(backupCount int) error {

	err := syscall.Chroot("/host")
	if err != nil {
		//log.Println("Failed to create chroot on /host\n\n\n")
		log.Println(err)
	}

	backupDir, err := getBackupDir(backupCount)
	certsDir := getCertificatesDir()
	kubeConfigs := getKubeConfigFiles()

	//stdout, err := exec.Command("/bin/bash", "-c", "chroot /host systemctl status etcd").Output()
	//if err != nil {
	//	log.Println(err)
	//}
	//fmt.Println(string(stdout))
	//
	//stdout, err = exec.Command("/bin/bash", "-c", "cp -r /etc/kubernetes/pki /opt/klovercloud/").Output()
	//if err != nil {
	//	log.Println(err)
	//}
	//
	//fmt.Println(string(stdout))

	err = CopyDirectory(certsDir, backupDir)
	if err != nil {
		fmt.Println(backupDir, certsDir, kubeConfigs)
		log.Println(err)
	}

	//for _, certFileName := range certificateList {
	//
	//	// add arguments
	//	cmd := exec.Command("cp")
	//	err = cmd.Run()
	//	if err != nil {
	//	}
	//}
	//for _, kubeConfigFile := range kubeConfigs {
	//}

	return err
}
