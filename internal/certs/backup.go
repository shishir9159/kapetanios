package certs

import (
	"context"
	"errors"
	"fmt"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	defer in.Close()
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

	//files, err := ioutil.ReadDir("testFolder")
	//	sfi, err := os.Stat(src)
	//	if err != nil {
	//		return
	//	}
	//	if !sfi.Mode().IsRegular() {
	//		// cannot copy non-regular files (e.g., directories,
	//		// symlinks, devices, etc.)
	//		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	//	}
	//	dfi, err := os.Stat(dst)
	//	if err != nil {
	//		if !os.IsNotExist(err) {
	//			return
	//		}
	//	} else {
	//		if !(dfi.Mode().IsRegular()) {
	//			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
	//		}
	//		if os.SameFile(sfi, dfi) {
	//			return
	//		}
	//	}

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

func BackupCertificatesKubeconfigs() {

	backupDir := getBackupDir(3)
	certsDir := getCertificatesDir()
	kubeConfigs := getKubeConfigFiles()

	err := syscall.Chroot("/host")
	if err != nil {

	}

	cmd := exec.Command("systemctl status etcd")
	err = cmd.Run()
	if err != nil {

	}

	println(backupDir, certsDir, kubeConfigs)

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
}
