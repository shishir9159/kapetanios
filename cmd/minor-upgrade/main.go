package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	maxAttempts = 3
	addr        = flag.String("addr", "kapetanios.default.svc.cluster.local:50051", "the address to connect to")
	retryPolicy = `{
		"methodConfig": [{
		  "retryPolicy": {
			  "MaxAttempts": 4,
			  "InitialBackoff": ".01s",
			  "MaxBackoff": ".01s",
			  "BackoffMultiplier": 4.0,
			  "RetryableStatusCodes": [ "UNAVAILABLE", "DEADLINE_EXCEEDED" ]
		  }
		}]}`
)

// todo:
//  yum repository for kubernetes
//  find in which repo the package belongs to
//  multiple repo handle

type Controller struct {
	ctx    context.Context
	log    zerolog.Logger
	distro string
}

func main() {

	// TODO: handle the following errors
	//  [azureuser@robi-infra-poc-1 ~]$ yum info kubectl
	//  Docker CE Stable - x86_64                                                                                                                                                         677 kB/s |  66 kB     00:00
	//  Kubernetes                                                                                                                                                                         47 kB/s |  33 kB     00:00
	//  Red Hat Enterprise Linux 8 for x86_64 - BaseOS from RHUI (RPMs)                                                                                                                   0.0  B/s |   0  B     00:00
	//  Errors during downloading metadata for repository 'rhel-8-for-x86_64-baseos-rhui-rpms':
	//    - Curl error (58): Problem with the local SSL certificate for https://rhui4-1.microsoft.com/pulp/repos/content/dist/rhel8/rhui/8/x86_64/baseos/os/repodata/repomd.xml [could not load PEM client certificate, OpenSSL error error:0200100D:system library:fopen:Permission denied, (no key found, wrong pass phrase, or wrong file format?)]
	//  Error: Failed to download metadata for repo 'rhel-8-for-x86_64-baseos-rhui-rpms': Cannot download repomd.xml: Cannot download repodata/repomd.xml: All mirrors were tried

	// yum info -b kubectl
	// rpm -q kubectl

	// TODO: optional backups

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	//if debug == true {
	//	out := os.Stdout
	//	logLevel := zerolog.DebugLevel
	//}

	logger := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339Nano,
		FormatLevel: func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("[%s]", i))
		},
		FormatMessage: func(i interface{}) string {
			return fmt.Sprintf("| %s |", i)
		},
		FormatCaller: func(i interface{}) string {
			return filepath.Base(fmt.Sprintf("%s", i))
		},
		PartsExclude: []string{
			zerolog.TimestampFieldName,
		},
	}).With().Timestamp().Caller().Stack().Logger()
	//.Level(zerolog.InfoLevel)

	// TODO:
	//  replace zap with zeroLog

	c := Controller{
		ctx: ctx,
		log: logger,
	}

	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(retryPolicy))
	if err != nil {
		c.log.Error().Err(err).
			Msg("couldn't connect to the kapetanios")

	}

	//grpc.WithDisableServiceConfig()
	defer func(conn *grpc.ClientConn) {
		er := conn.Close()
		if er != nil {
			c.log.Error().Err(er).
				Msg("failed to close the grpc connection")
		}
	}(conn)

	var distro string

	for i := 0; i < maxAttempts; i++ {
		var skip bool
		skip, distro, err = Prerequisites(c, conn)
		if err != nil {
			c.log.Error().Err(err).
				Int("attempt", i).
				Msg("failed to check prerequisites for cluster upgrade")
		}

		if skip {
			break
		}
	}

	c.distro = distro

	var version string

	for i := 0; i < maxAttempts; i++ {
		var skip bool
		skip, version, err = availableVersions(c, conn)

		if err != nil {
			c.log.Error().Err(err).
				Int("attempt", i).
				Msg("failed to fetch minor versions for the kubernetes upgrade")
		}

		if skip {
			break
		}
	}

	// todo: include in the testing
	testing := false
	version = "1.29.0-150500.1.1"
	//version = "1.26.6-1.1"

	if testing {
		//version = kubernetesVersion
	}

	// TODO: refactor
	//   plan := "v1.26.6"
	//   clusterVersion should be separated from package(+) or components(-) version?

	var diff string
	for i := 0; i < maxAttempts; i++ {
		var skip bool
		skip, diff, err = compatibility(c, "v1.29.0", conn)
		if err != nil {
			c.log.Error().Err(err).
				Int("attempt", i).
				Msg("failed to diff")
		}

		if skip {
			break
		}
	}

	c.log.Info().
		Str("version", version).
		Str("diff", diff).
		Msg("diff for upgrade plan")

	for i := 0; i < maxAttempts; i++ {
		skip, er := k8sComponentsUpgrade(c, "kubeadm", version, conn)
		if er != nil {
			c.log.Error().Err(er).
				Int("attempt", i).
				Msg("failed to upgrade kubeadm")
		}

		if skip {
			break
		}
	}

	var plan string
	for i := 0; i < maxAttempts; i++ {
		var skip bool
		skip, plan, err = upgradePlan(c, conn)
		if err != nil {
			c.log.Error().Err(err).
				Int("attempt", i).
				Msg("failed to get upgrade plan")
		}

		if skip {
			break
		}
	}

	fmt.Println(plan)

	for i := 0; i < maxAttempts; i++ {
		skip, er := clusterUpgrade(c, version, conn)
		if er != nil {
			c.log.Error().Err(er).
				Int("attempt", i).
				Msg("failed to upgrade cluster")
		}

		if skip {
			break
		}
	}

	for i := 0; i < maxAttempts; i++ {
		skip, er := k8sComponentsUpgrade(c, "kubelet", version, conn)
		if er != nil {
			c.log.Error().Err(er).
				Int("attempt", i).
				Msg("failed to upgrade kubelet")
		}

		if skip {
			break
		}
	}

	for i := 0; i < maxAttempts; i++ {
		skip, er := restartComponent(c, "kubelet", conn)
		if er != nil {
			c.log.Error().Err(er).
				Int("attempt", i).
				Msg("failed to restart kubelet")
		}

		if skip {
			break
		}
	}

	// TODO: etcd node restart based
	//  on condition

	for i := 0; i < maxAttempts; i++ {
		skip, er := k8sComponentsUpgrade(c, "kubectl", version, conn)
		if er != nil {
			c.log.Error().Err(er).
				Int("attempt", i).
				Msg("failed to upgrade kubectl")
		}

		if skip {
			break
		}
	}

	// TODO: sanityCheck/validation for cluster upgrade & finalizer
}
