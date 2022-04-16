package infra

import (
	"zoo/config"
	"github.com/go-redis/redis/v8"
	"log"
)

var RDB *redis.Client

func InitRedis() {
	c := config.Config
	RDB = redis.NewClient(&redis.Options{
		Addr:     c.Redis.Address,
		Password: c.Redis.Password,
		DB:       c.Redis.Database,
	})
	log.Printf("InitRedis success")
}
