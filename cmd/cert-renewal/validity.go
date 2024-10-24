package main

import (
	"crypto/x509"
	"fmt"
	"log"
	"sync"
	"time"
)

// TODO: validity check only pods

// TODO: if "externally managed" value shows yes
//  suggestions:
//  step 1. cordon, drain, delete: kubectl drain <node-name> --ignore-daemonsets --delete-local-data; kubectl delete node <node-name>
//  step 2. kubeadm token create --print-join-command --config /etc/kubernetes/kubeadm/kubeadm-config.yaml
//  step 3. kubeadm init phase upload-certs --upload-certs --config /etc/kubernetes/kubeadm/kubeadm-config.yaml
//  step 4. kubeadm join <master-node>:6443 --token <23-characters-long-token>
//    --discovery-token-ca-cert-hash sha256:<64-characters-long-token> --control-plane --certificate-key
//   <64-characters-long-certificate-from-the-output-of-step-3> --apiserver-advertise-address <master-node-ip> --v=14

// WARNING: for externally managed being true

var (
	certPeriodValidationMutex  sync.Mutex
	certPeriodValidationCached = map[string]struct{}{}
)

func ValidateCertPeriod(cert *x509.Certificate, offset time.Duration) error {
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

func CheckCertificateValidity(baseName string, cert *x509.Certificate) {
	certPeriodValidationMutex.Lock()
	defer certPeriodValidationMutex.Unlock()
	if _, exists := certPeriodValidationCached[baseName]; exists {
		return
	}
	certPeriodValidationCached[baseName] = struct{}{}

	if err := ValidateCertPeriod(cert, 0); err != nil {
		log.Println(fmt.Errorf("WARNING: could not validate bounds for certificate %s: %v", baseName, err))
	}
}
