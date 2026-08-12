package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	// RedisURL is an optional full connection URL (e.g. Upstash rediss://... TCP).
	// When non-empty it takes priority and all other fields are ignored.
	RedisURL       string
	Addr           string
	Password       string
	DB             int
	Protocol       int
	MaxIdleConns   int
	MaxActiveConns int
}

func NewCacheStorage(config RedisConfig) (*redis.Client, error) {
	ctx := context.Background()

	var opt *redis.Options

	if config.RedisURL != "" {
		// Upstash / any Redis URL (rediss:// for TLS, redis:// for plain TCP)
		parsed, err := redis.ParseURL(config.RedisURL)
		if err != nil {
			return nil, fmt.Errorf("redis: parse url: %w", err)
		}
		opt = parsed
	} else {
		// Local / self-hosted Redis via individual fields
		opt = &redis.Options{
			Addr:     config.Addr,
			Password: config.Password,
			DB:       config.DB,
			Protocol: config.Protocol,
		}
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
