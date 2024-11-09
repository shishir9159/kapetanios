package main

import (
	"bytes"
	"crypto/x509"
	"regexp"
	"strings"

	"fmt"
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"go.uber.org/zap"
	"log"
	"os/exec"
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

func replaceConsecutiveSpaces(s string) string {
	// Regular expression to match two or more consecutive spaces
	re := regexp.MustCompile(`\s{2,}`)
	// Replace matches with "-"
	return re.ReplaceAllString(s, "+")
}

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

func certExpiration(c Controller, connection pb.ValidityClient) (time.Time, time.Time, error) {
	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Error("Failed to create chroot on /host",
			zap.Error(err))
		return time.Time{}, time.Time{}, err
	}

	cmd := exec.Command("/bin/bash", "-c", "kubeadm certs check-expiration --config /etc/kubernetes/kubeadm-config.yaml")

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdoutBuf, &stderrBuf

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
		c.log.Error("Failed to check cert expiration date",
			zap.Error(err))
		return time.Time{}, time.Time{}, err
	}

	if err = changedRoot(); err != nil {
		c.log.Fatal("Failed to exit from the updated root",
			zap.Error(err))
	}

	outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())

	// todo: string operations

	tillCertsSubstr := "CERTIFICATE                EXPIRES                  RESIDUAL TIME   CERTIFICATE AUTHORITY   EXTERNALLY MANAGED\n"
	tillCaAuthoritiesSubstr := "CERTIFICATE AUTHORITY   EXPIRES                  RESIDUAL TIME   EXTERNALLY MANAGED\n"

	indexCerts := strings.Index(outStr, tillCertsSubstr)
	indexCaAuthorities := strings.Index(outStr, tillCaAuthoritiesSubstr)

	certsString := outStr[indexCerts+len(tillCertsSubstr) : indexCaAuthorities]
	caAuthoritiesString := outStr[indexCaAuthorities+len(tillCaAuthoritiesSubstr):]

	certs := strings.Split(certsString, "\n")
	caAuthorities := strings.Split(caAuthoritiesString, "\n")

	var certificates []*pb.Certificate

	for _, cert := range certs {
		cert = replaceConsecutiveSpaces(cert)
		fmt.Println(cert)
		fields := strings.Split(cert, "+")

		// to skip the last empty lines from the certs
		if len(fields) != 6 {
			break
		}

		certificate := pb.Certificate{
			Name:                 fields[0],
			Expires:              fields[1],
			ResidualTime:         fields[2],
			CertificateAuthority: fields[3],
			ExternallyManaged:    fields[4],
		}

		certificates = append(certificates, &certificate)
	}

	var certificateAuthorities []*pb.CertificateAuthority

	for _, ca := range caAuthorities {
		ca = replaceConsecutiveSpaces(ca)
		fields := strings.Split(ca, "+")

		if len(fields) != 6 {
			break
		}

		certificateAuthority := pb.CertificateAuthority{
			Name:              fields[0],
			Expires:           fields[1],
			ResidualTime:      fields[2],
			ExternallyManaged: fields[3],
		}

		certificateAuthorities = append(certificateAuthorities, &certificateAuthority)
	}

	c.log.Info("outString and errString",
		zap.String("outStr", outStr),
		zap.String("errStr", errStr))

	rpc, err := connection.ExpirationInfo(c.ctx,
		&pb.Expiration{
			ValidCertificate:       false,
			Certificates:           certificates,
			CertificateAuthorities: certificateAuthorities,
		})

	if err != nil {
		c.log.Error("could not send status update: ", zap.Error(err))
	}

	c.log.Info("Status Update",
		zap.Bool("response received", rpc.GetReceived()))

	return time.Time{}, time.Time{}, err
}
