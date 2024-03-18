package cache

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	Client *redis.Client
}

func (c *RedisCache) ConnectCache() {
	addr := os.Getenv("CACHE_ADDR")

	opt, err := redis.ParseURL(addr)
	if err != nil {
		log.Fatal("Incorrect CACHE_ADDR URL provided")
	}

	client := redis.NewClient(opt)

	if status := client.Ping(context.Background()); status.Err() != nil {
		log.Fatal("Connection Refused", status.Err())
	}

	c.Client = client
}
