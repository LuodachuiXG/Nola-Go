package db

import (
	"context"
	"nola-go/internal/config"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(cfg *config.Config) (*redis.Client, error) {
	cli := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// ping
	err := cli.Ping(context.Background()).Err()
	if err != nil {
		return nil, err
	}
	return cli, nil
}
