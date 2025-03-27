package config

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
)

var redisDbForWrite *redis.Client
var redisDbForRead *redis.Client
var RedisDbForPermissions *redis.Client

func InitRedisWriteConnection() error {
	redisDbForWrite = redis.NewClient(&redis.Options{
		Addr:     RedisServerForWrite,
		Password: RedisServerPassword,
		DB:       0, // use default DB
	})
	//log.Logger.Info("[INFO] Redis Write connection initialized")
	//_, err := redisDbForWrite.Ping(context.Background()).Result()
	if err != nil {
		//log.Logger.Fatal(err)
	}
	var m map[string]int
	m = make(map[string]int)
	m["/api/"] = 1
	jsonString, err := json.Marshal(m)
	//if err != nil {
	//	log.Logger.Error(err.Error())
	//}
	err = redisDbForWrite.Set(context.Background(), "test", jsonString, 0).Err()

	va, err := redisDbForWrite.Get(context.Background(), "test").Result()
	if err != nil {
		//log.Logger.Error(err.Error())
	}
	log.Logger.Info("printing here:", va)

	sec := map[string]int{}
	if err := json.Unmarshal([]byte(va), &sec); err != nil {
		panic(err)
	}

	//log.Logger.Info("output map: ", "val", sec)
	//log.Logger.Info("from map va key", "/api/", sec["/api/"])
	return nil
}

func InitRedisReadConnection() error {
	redisDbForRead = redis.NewClient(&redis.Options{
		Addr:     RedisServerForRead,
		Password: RedisServerPassword, // no password set
		DB:       0,                   // use default DB
	})

	//log.Logger.Info("[INFO] Redis Read connection initialized")
	return nil
}

func InitRedisPermissionsConnection() error {
	RedisDbForPermissions = redis.NewClient(&redis.Options{
		Addr:     RedisAddress,
		Password: RedisPassword,
		DB:       0, // use default DB
	})

	err := RedisDbForPermissions.Ping(context.Background()).Err()
	if err != nil {
		//log.Logger.Error(err)
	}

	return nil
}
