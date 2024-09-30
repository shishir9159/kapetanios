package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func getCertificatesDir() string {
	//	will be called after getKubeadmFileLocation
	// default:	/etc/kubernetes/pki

	client, err := orchestration.NewClient()
	if err != nil {
		fmt.Println("Error creating Kubernetes client: %v", err)
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
// the same, then return success. Otherise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {

	in, err := os.Open(src)
	if err != nil {
		return
	}

	defer func(in *os.File) {
		err := in.Close()
		if err != nil {

		}
	}(in)

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func getBackupDir(backupCount int) string {

	baseDir := "/opt/klovercloud"
	//backupDirPattern := "cert-backup-*"
	backupDirPattern := "certs-backup-"
	err := syscall.Chroot("/host")
	if err != nil {
		return ""
	}

	//files, err := os.ReadDir("")
	//sfi, err := os.Stat(src)
	//if err != nil {
	//	return ""
	//}
	//if !sfi.Mode().IsRegular() {
	//	// cannot copy non-regular files (e.g., directories,
	//	// symlinks, devices, etc.)
	//	return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	//}
	//dfi, err := os.Stat(dst)
	//if err != nil {
	//	if !os.IsNotExist(err) {
	//		return ""
	//	}
	//} else {
	//	if !(dfi.Mode().IsRegular()) {
	//		return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
	//	}
	//	if os.SameFile(sfi, dfi) {
	//		return
	//	}
	//}

	if _, err = os.Stat(baseDir); err == nil {
		// path/to/whatever exists

	} else if errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir(baseDir, 600)

	}

	// TO-DO: handle possible permission errors
	if err != nil {
		return ""
	}

	// check if it works with folder
	oldBackupFiles, err := filepath.Glob(filepath.Join(baseDir, backupDirPattern))
	if err != nil {
		// non match
	}

	// to-do: apply the old logic
	if len(oldBackupFiles) > backupCount {
		err = os.Mkdir("/opt/klovercloud/certs-backup-1", 600)
		return "/opt/klovercloud/certs-backup-1"
	}

	return "/opt/klovercloud/certs-backup-1"
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

	backupDir := getBackupDir(backupCount)
	certsDir := getCertificatesDir()
	kubeConfigs := getKubeConfigFiles()

	//err := syscall.Chroot("/host")
	//if err != nil {
	//	log.Println("Failed to create chroot on /host")
	//	log.Println(err)
	//}

	cmd, err := exec.Command("/bin/bash", "-c", "chrtoot /host systemctl status etcd").Output()
	if err != nil {
		log.Println(err)
	}

	fmt.Println(string(cmd))

	//err = CopyDirectory(certsDir, backupDir)
	//if err != nil {
	fmt.Println(backupDir, certsDir, kubeConfigs)
	//	log.Fatalln(err)
	//	return err
	//}

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
