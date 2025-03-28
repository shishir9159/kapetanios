package config

import (
	"k8s.io/client-go/tools/cache"
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

var isK8 = "False"

var KubeConfigFile = "config"

var AuthServerBaseURl = ""
var AuthServerTokenValidateEndpoint = ""
var MultiClusterBaseURL = ""
var MultiClusterClusterPermissionCheckEndpoint = ""
var CacheTtlMin = 60 * time.Minute
var PageLimit = int64(10)
var RedisServerForWrite string
var RedisServerForRead string
var RedisServerPassword string
var ConnectRedis string
var RedisAddress string
var RedisPassword string

func GetRunMode() string {
	return RunMode
}

func LoadEnvironmentVariables() {
	RunMode = os.Getenv("RUN_MODE")
	if RunMode == "" {
		RunMode = DEVELOP
	}

	log.Logger.Infow("RUN MODE:", "value", RunMode)

	if RunMode != PRODUCTION {
		//Load .env file
		err := godotenv.Load()
		if err != nil {
			log.Logger.Errorw("Failed to Load environment file", "err", err.Error())
			os.Exit(1)
		}
	} else {
		err := InitRedisPermissionsConnection()
		if err != nil {
			log.Logger.Error(err)
		}
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
	initCacheStore()
}

func IsK8() string {
	return isK8
}

func GetAuthServerValidationEndpoint() string {
	return AuthServerBaseURl + AuthServerTokenValidateEndpoint
}

func GetMultiClusterPermissionEndpoint(clusterId string) string {
	return MultiClusterBaseURL + MultiClusterClusterPermissionCheckEndpoint + clusterId
}

func EnvironmentVariablesChecker() error {
	if AuthServerTokenValidateEndpoint == "" || AuthServerBaseURl == "" || MultiClusterClusterPermissionCheckEndpoint == "" || MultiClusterBaseURL == "" {
		errorMsg := ""
		if AuthServerBaseURl == "" {
			errorMsg += "auth server base url, "
		}
		if AuthServerTokenValidateEndpoint == "" {
			errorMsg += "auth server token endpoint, "
		}
		if MultiClusterBaseURL == "" {
			errorMsg += "multi cluster base url, "
		}
		if MultiClusterClusterPermissionCheckEndpoint == "" {
			errorMsg += "multi cluster permission endpoint, "
		}
		errorMsg = errorMsg[:len(errorMsg)-2] + " values missing"
		log.Logger.Errorw("missing env values", "err", errorMsg)
	}
	return nil
}

var cacheStore cache.Store

type KeyValue struct {
	Key   string
	Value interface{}
}

// cacheKeyFunc defines the key function required in TTLStore.
func cacheKeyFunc(obj interface{}) (string, error) {
	return obj.(KeyValue).Key, nil
}

func initCacheStore() {
	cacheStore = cache.NewTTLStore(cacheKeyFunc, CacheTtlMin)
}
func GetCacheStore() cache.Store {
	return cacheStore
}
