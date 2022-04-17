package infra

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
	"git.garena.com/xinlong.wu/zoo/config"
)

var RedisClient *redis.Client

func InitRedis() {
	c := config.Config
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     c.Redis.Address,
		Password: c.Redis.Password,
		DB:       c.Redis.Database,
	})
	log.Printf("InitRedis success")
}

var RedisUtil = &redisUtil{}

type redisUtil struct {
}

func (rds *redisUtil) GetOrSet(key string, load func(key string, ret any) error, expiration time.Duration, ret any) error {
	result, err := RedisClient.Get(context.Background(), key).Result()
	if err == redis.Nil {
		// not found, load and set and return
		err := load(key, ret)
		if err != nil {
			return err
		}
		if err = rds.SetEX(key, ret, expiration); err != nil {
			return err
		}
		return nil
	}

	// found it, unmarshal and return
	if err = json.Unmarshal([]byte(result), ret); err != nil {
		return err
	}
	return nil
}

func (rds *redisUtil) SetEX(key string, value any, expiration time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = RedisClient.SetEX(context.Background(), key, string(bytes), expiration).Result()
	return err
}
