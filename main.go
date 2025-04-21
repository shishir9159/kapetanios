package main

import (
	"fmt"
	"regexp"
	"strings"
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

func main() {
	outStr := "[check-expiration] Reading configuration from the cluster...\n[check-expiration] FYI: You can look at this config file with 'kubectl -n kube-system get cm kubeadm-config -o yaml'\nW0421 22:10:12.235818  297192 configset.go:78] Warning: No kubeproxy.config.k8s.io/v1alpha1 config is loaded. Continuing without it: configmaps \"kube-proxy\" not found\n\nCERTIFICATE                EXPIRES                  RESIDUAL TIME   CERTIFICATE AUTHORITY   EXTERNALLY MANAGED\nadmin.conf                 Apr 20, 2026 15:01 UTC   363d            ca                      no      \napiserver                  Apr 20, 2026 15:01 UTC   363d            ca                      no      \napiserver-kubelet-client   Apr 20, 2026 15:01 UTC   363d            ca                      no      \ncontroller-manager.conf    Apr 20, 2026 15:01 UTC   363d            ca                      no      \nfront-proxy-client         Apr 20, 2026 15:01 UTC   363d            front-proxy-ca          no      \nscheduler.conf             Apr 20, 2026 15:01 UTC   363d            ca                      no      \nsuper-admin.conf           Apr 20, 2026 15:01 UTC   363d            ca                      no      \n\nCERTIFICATE AUTHORITY   EXPIRES                  RESIDUAL TIME   EXTERNALLY MANAGED\nca                      Apr 18, 2035 15:01 UTC   9y              no      \nfront-proxy-ca          Apr 18, 2035 15:01 UTC   9y              no      \n"
	tillCertsSubstr := []string{"CERTIFICATE", "EXPIRES", "RESIDUAL TIME", "CERTIFICATE AUTHORITY", "EXTERNALLY MANAGED"}
	tillCaAuthoritiesSubstr := []string{"CERTIFICATE AUTHORITY", "EXPIRES", "RESIDUAL TIME", "EXTERNALLY MANAGED"}

	indexCerts := indexTillLineWithWordsInOrder(outStr, tillCertsSubstr)
	indexCaAuthorities := indexTillLineWithWordsInOrder(outStr, tillCaAuthoritiesSubstr)

	//indexCerts := strings.Index(strings.ReplaceAll(outStr, " ", ""), strings.ReplaceAll(tillCertsSubstr, " ", ""))
	//indexCaAuthorities := strings.Index(strings.ReplaceAll(outStr, " ", ""), strings.ReplaceAll(tillCaAuthoritiesSubstr, " ", ""))

	fmt.Println(indexCerts, indexCaAuthorities)

	//indexCerts += strings.Index(outStr[indexCerts:], "\n") + 1
	//indexCaAuthorities += strings.Index(outStr[indexCaAuthorities:], "\n") + 1

	certsString := outStr[indexCerts:indexCaAuthorities]
	caAuthoritiesString := outStr[indexCaAuthorities:]

	certs := strings.Split(certsString, "\n")
	caAuthorities := strings.Split(caAuthoritiesString, "\n")

	for _, cert := range certs {
		cert = replaceConsecutiveSpaces(cert)
		fields := strings.Split(cert, "+")

		fmt.Println(fields)

		// to skip the last empty lines from the certs
		// TODO:
		if len(fields) != 6 {
			break
		}
	}

	for _, ca := range caAuthorities {
		ca = replaceConsecutiveSpaces(ca)
		fields := strings.Split(ca, "+")

		if len(fields) != 5 {
			break
		}

		fmt.Println(fields)
	}
}
