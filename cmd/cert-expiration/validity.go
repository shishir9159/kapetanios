package main

import (
	"bytes"
	"crypto/x509"
	"fmt"
	"github.com/shishir9159/kapetanios/utils"
	"go.uber.org/zap"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// TODO: validity check only pods

// TODO: if "externally managed" value shows yes
//  suggestions:
//  step 1. cordon, drain, delete: kubectl drain <node-name> --ignore-daemonsets --delete-local-data; kubectl delete node <node-name>
//  step 2. kubeadm token create --print-join-command --config /etc/kubernetes/kubeadm-config.yaml
//  step 3. kubeadm init phase upload-certs --upload-certs --config /etc/kubernetes/kubeadm-config.yaml
//  step 4. kubeadm join <master-node>:6443 --token <23-characters-long-token>
//    --discovery-token-ca-cert-hash sha256:<64-characters-long-token> --control-plane --certificate-key
//   <64-characters-long-certificate-from-the-output-of-step-3> --apiserver-advertise-address <master-node-ip> --v=14

// should it be with certs renewal and minor-upgrade?

var (
	certPeriodValidationMutex  sync.Mutex
	certPeriodValidationCached = map[string]struct{}{}
)

func validateCertPeriod(cert *x509.Certificate, offset time.Duration) error {
	period := fmt.Sprintf("NotBefore: %v, NotAfter: %v", cert.NotBefore, cert.NotAfter)
	now := time.Now().Add(offset)
	if now.Before(cert.NotBefore) {
		return fmt.Errorf("the certificate is not valid yet: %s", period)
	}
	if now.After(cert.NotAfter) {
		return fmt.Errorf("the certificate has expired: %s", period)
	}
	return nil
}

func checkCertificateValidity(baseName string, cert *x509.Certificate) {
	certPeriodValidationMutex.Lock()
	defer certPeriodValidationMutex.Unlock()
	if _, exists := certPeriodValidationCached[baseName]; exists {
		return
	}
	certPeriodValidationCached[baseName] = struct{}{}

	if err := validateCertPeriod(cert, 0); err != nil {
		log.Println(fmt.Errorf("WARNING: could not validate bounds for certificate %s: %v", baseName, err))
	}
}

func certExpiration(log *zap.Logger) (time.Time, time.Time, error) {
	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		log.Error("Failed to create chroot on /host",
			zap.Error(err))
		return time.Time{}, time.Time{}, err
	}

	var stdoutBuf, stderrBuf bytes.Buffer

	cmd := exec.Command("/bin/bash", "-c", "kubeadm certs check-expiration --config /etc/kubernetes/kubeadm-config.yaml")
	err = cmd.Run()

	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err = cmd.Run()

	// TODO: format the following
	//CERTIFICATE                EXPIRES                  RESIDUAL TIME   CERTIFICATE AUTHORITY   EXTERNALLY MANAGED
	//admin.conf                 Oct 18, 2025 04:31 UTC   350d            ca                      no
	//apiserver                  Oct 18, 2025 04:31 UTC   350d            ca                      no
	//apiserver-kubelet-client   Oct 18, 2025 04:31 UTC   350d            ca                      no
	//controller-manager.conf    Oct 18, 2025 04:31 UTC   350d            ca                      no
	//front-proxy-client         Oct 18, 2025 04:31 UTC   350d            front-proxy-ca          no
	//scheduler.conf             Oct 18, 2025 04:31 UTC   350d            ca                      no
	//
	//CERTIFICATE AUTHORITY   EXPIRES                  RESIDUAL TIME   EXTERNALLY MANAGED
	//ca                      May 27, 2034 13:31 UTC   9y              no
	//front-proxy-ca          May 27, 2034 13:31 UTC   9y              no

	if err != nil {
		log.Error("Failed to check cert expiration date",
			zap.Error(err))
		return time.Time{}, time.Time{}, err
	}

	outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())

	checkingSubstr := strings.Contains(outStr, "CERTIFICATE                EXPIRES                  RESIDUAL TIME   CERTIFICATE AUTHORITY   EXTERNALLY MANAGED")

	log.Info("outString and errString",
		zap.String("outStr", outStr),
		zap.String("errStr", errStr),
		zap.Bool("checkingSubstr", checkingSubstr))

	if err = changedRoot(); err != nil {
		log.Fatal("Failed to exit from the updated root",
			zap.Error(err))
	}

	return time.Time{}, time.Time{}, err

}