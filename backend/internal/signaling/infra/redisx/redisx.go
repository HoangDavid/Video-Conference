package redisx

import (
	"context"
	"sync"
	"time"
	"vidcall/pkg/logger"

	goredis "github.com/redis/go-redis/v9"
)

var (
	once   sync.Once
	client *goredis.Client
)

func Init(addr, password string, db int) {
	once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		defer cancel()

		client = goredis.NewClient(&goredis.Options{
			Addr:        addr,
			Password:    password,
			DB:          db,
			DialTimeout: 5 * time.Second,
		})

		log := logger.GetLog(ctx).With("layer", "infra", "service", "redis")
		err := client.Ping(ctx).Err()
		if err != nil {
			log.Error("Unable to connect to Redis")
			return
		}
	})

}

func C() *goredis.Client {
	return client
}
