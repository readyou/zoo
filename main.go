package main

import (
	"context"
	"fmt"
	infra "git.garena.com/xinlong.wu/zoo/infra"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

func main() {
	infra.InitDB()
	infra.InitRedis()
	now, err := infra.DB.QueryString("select now()")
	if err == nil {
		log.Printf("mysql: now: %s\n", now)
	}

	ctx := context.Background()
	val, err := infra.RDB.Set(ctx, "hello", "world", time.Second).Result()
	val, err = infra.RDB.Get(ctx, "hello").Result()
	switch {
	case err == redis.Nil:
		fmt.Println("key does not exist")
	case err != nil:
		fmt.Println("Get failed", err)
	case val == "":
		fmt.Println("value is empty")
	default:
		log.Println("redis: key=hello, value=", val)
	}

}
