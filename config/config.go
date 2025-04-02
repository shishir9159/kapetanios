package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// todo: debug mode
// todo: dev mode
// todo: prod mode

var RunMode string
var ServerPort string

func LoadEnvironmentVariables() {

	RunMode = os.Getenv("RUN_MODE")
	if RunMode == "" {
		RunMode = DEVELOP
	}

	//log.Logger.Infow("RUN MODE:", "value", RunMode)

	if RunMode != PRODUCTION {
		//Load .env file
		//err := godotenv.Load()
		//if err != nil {
		//	log.Logger.Errorw("Failed to Load environment file", "err", err.Error())
		//	os.Exit(1)
		//}
	} else {
		//err := InitRedisPermissionsConnection()
		//if err != nil {
		//	log.Logger.Error(err)
		//}
	}
	ServerPort = os.Getenv("SERVER_PORT")
	isK8 = os.Getenv("IS_K8")
	KubeConfigFile = os.Getenv("KUBE_CONFIG_FILE")
	AuthServerBaseURl = os.Getenv("AUTH_SERVER_BASE_URL")
	AuthServerTokenValidateEndpoint = os.Getenv("AUTH_SERVER_VALIDATE_TOKEN_ENDPOINT")
	MultiClusterBaseURL = os.Getenv("MULTI_CLUSTER_GATEWAY_BASE_URL")
	MultiClusterClusterPermissionCheckEndpoint = os.Getenv("MULTI_CLUSTER_GATEWAY_CLUSTER_PERMISSION_CHECK_ENDPOINT")
	t, _ := strconv.Atoi(os.Getenv("CACHE_TTL_IN_MIN"))
	CacheTtlMin = time.Duration(t) * time.Minute
	pl, _ := strconv.Atoi(os.Getenv("PAGE_LIMIT"))
	PageLimit = int64(pl)
	ConnectRedis = strings.TrimSpace(os.Getenv("CONNECT_REDIS"))
	RedisServerForWrite = strings.TrimSpace(os.Getenv("REDIS_SERVER_FOR_WRITE"))
	RedisServerForRead = strings.TrimSpace(os.Getenv("REDIS_SERVER_FOR_READ"))
	RedisServerPassword = strings.TrimSpace(os.Getenv("REDIS_SERVER_PASSWORD"))
	RedisAddress = strings.TrimSpace(os.Getenv("REDIS_ADDRESS"))
	RedisPassword = strings.TrimSpace(os.Getenv("REDIS_PWD"))
	LoadTemporalVariables()
	initCacheStore()
}
