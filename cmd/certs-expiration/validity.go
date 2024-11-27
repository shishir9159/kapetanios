package main

import (
	"bytes"
	"crypto/x509"
	"fmt"
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

// TODO: validity check only pods

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
		c.log.Fatal().Err(err).
			Msg("Failed to create chroot on /host")
	}

	// TODO: when is the config file necessary?
	cmd := exec.Command("/bin/bash", "-c", "kubeadm certs check-expiration --config /etc/kubernetes/kubeadm/kubeadm-config.yaml")

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdoutBuf, &stderrBuf

	err = cmd.Run()

	if err != nil {
		c.log.Error().Err(err).
			Msg("Failed to check cert expiration date")
	}

	if err = changedRoot(); err != nil {
		c.log.Fatal().Err(err).
			Msg("Failed to exit from the updated root")

	}

	outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())

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

		if len(fields) != 5 {
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

	c.log.Info().
		Str("outStr", outStr).
		Str("errStr", errStr).
		Send()

	rpc, err := connection.ExpirationInfo(c.ctx,
		&pb.Expiration{
			ValidCertificate:       true, //tOdO: count
			Certificates:           certificates,
			CertificateAuthorities: certificateAuthorities,
		})

	if err != nil {
		c.log.Error().Err(err).
			Msg("could not send status update")
	}

	c.log.Info().
		Bool("response received", rpc.ResponseReceived).
		Msg("status update")

	return time.Time{}, time.Time{}, err
}
