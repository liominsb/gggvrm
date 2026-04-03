package config

import (
	"gggvrm/global"
	"log"

	"github.com/go-redis/redis"
)

func initRedis() {
	RedisCilnet := redis.NewClient(&redis.Options{
		Addr:     Appconf.Database.Addr,
		Password: Appconf.Database.Password, // no password set
		DB:       0,                         // use default DB
	})

	_, err := RedisCilnet.Ping().Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	global.RedisDB = RedisCilnet
}
