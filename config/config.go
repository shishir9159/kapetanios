package config

import (
	"os"
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
}
