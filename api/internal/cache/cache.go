package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr           string
	Password       string
	DB             int
	Protocol       int
	MaxIdleConns   int
	MaxActiveConns int
}

func NewCacheStorage(config RedisConfig) (*redis.Client, error) {
	ctx := context.Background()

	opt := &redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
		Protocol: config.Protocol,
	}

	if config.MaxIdleConns > 0 {
		opt.MaxIdleConns = config.MaxIdleConns
	}
	if config.MaxActiveConns > 0 {
		opt.MaxActiveConns = config.MaxActiveConns
	}

	rdb := redis.NewClient(opt)

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return rdb, nil
}
