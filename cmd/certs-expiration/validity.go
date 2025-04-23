package main

import (
	"bytes"
	"crypto/x509"
	"fmt"
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"log"
	"os"
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
	// TODO: sanity check
	// to match two or more consecutive spaces
	re := regexp.MustCompile(`\s{2,}`)
	// Replace matches with "-"
	return re.ReplaceAllString(s, "+")
}

func indexTillLineWithWordsInOrder(text string, words []string) int {
	pattern := strings.Join(words, `.*?`) // Non-greedy match between words
	re := regexp.MustCompile(pattern)

	startIdx := 0
	lines := strings.SplitAfter(text, "\n")

	for _, line := range lines {
		startIdx += len(line)
		if re.MatchString(line) {
			return startIdx
		}
	}

	return -1
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

func missingCertificate(s string) []string {

	re := regexp.MustCompile(`!MISSING!\s+([^\s]+)`)
	matches := re.FindAllStringSubmatch(s, -1)

	var missingCerts []string
	for _, match := range matches {
		if len(match) > 1 {
			missingCerts = append(missingCerts, match[1])
		}
	}

	return missingCerts
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

	// TODO: in memory config file generation from the configMap
	// TODO: when is the config file necessary?
	cmd := exec.Command("/bin/bash", "-c", "kubeadm certs check-expiration")
	//cmd := exec.Command("/bin/bash", "-c", "kubeadm certs check-expiration --config /etc/kubernetes/kubeadm/kubeadm-config.yaml")

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

	tillCertsSubstr := []string{"CERTIFICATE", "EXPIRES", "RESIDUAL TIME", "CERTIFICATE AUTHORITY", "EXTERNALLY MANAGED"}
	tillCaAuthoritiesSubstr := []string{"CERTIFICATE AUTHORITY", "EXPIRES", "RESIDUAL TIME", "EXTERNALLY MANAGED"}

	indexCerts := indexTillLineWithWordsInOrder(outStr, tillCertsSubstr)
	indexCaAuthorities := indexTillLineWithWordsInOrder(outStr, tillCaAuthoritiesSubstr)

	certsString := outStr[indexCerts:indexCaAuthorities]
	caAuthoritiesString := outStr[indexCaAuthorities:]

	missingCerts := missingCertificate(certsString)
	if missingCerts != nil {
		//	TODO:
		//	 separate message for missing certs

		//	remove the lines containing Missing

		//	TODO:
		//	 update indexCaAuthorities and move caAuthoritiesString to a different place
	}

	certs := strings.Split(certsString, "\n")
	caAuthorities := strings.Split(caAuthoritiesString, "\n")

	var certificates []*pb.Certificate

	for _, cert := range certs {
		cert = replaceConsecutiveSpaces(cert)
		fields := strings.Split(cert, "+")

		c.log.Info().
			Str("cert", cert).
			Msg("certificates split by +")

		// to skip the last empty lines from the certs
		// TODO:
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
		Msg("certificate expiration")

	rpc, err := connection.ExpirationInfo(c.ctx,
		&pb.Expiration{
			ValidCertificate:       true, //tOdO: count
			NodeName:               os.Getenv("NODE_NAME"),
			MissingCertificates:    missingCerts,
			Certificates:           certificates,
			CertificateAuthorities: certificateAuthorities,
		})

	if err != nil {
		c.log.Error().Err(err).
			Msg("could not send status update")
	}

	c.log.Info().
		Bool("response received", rpc.GetResponseReceived()).
		Msg("status update")

	return time.Time{}, time.Time{}, err
}
