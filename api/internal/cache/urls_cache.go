package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/DanielJohn17/url-shortner/api/internal/config"
	"github.com/redis/go-redis/v9"
)

type URLCacheRepositoryInt interface {
	GetUrl(ctx context.Context, shortUrl string) (*UrlCache, error)
	SetUrl(ctx context.Context, url *UrlCache) error
}

type URLCacheRepository struct {
	rdb *redis.Client
}

type UrlCache struct {
	ShortUrl  string    `redis:"short_url"`
	LongUrl   string    `redis:"long_url"`
	UsedCount int       `redis:"used_count"`
	CreatedAt time.Time `redis:"created_at"`
}

func NewUrlCacheRepository(rdb *redis.Client) *URLCacheRepository {
	return &URLCacheRepository{
		rdb: rdb,
	}
}

func (c *URLCacheRepository) GetUrl(ctx context.Context, shortUrl string) (*UrlCache, error) {
	var urlCache UrlCache

	cmd := c.rdb.HGetAll(ctx, shortUrl)

	fields, err := cmd.Result()
	if err != nil {
		return nil, fmt.Errorf("Url not found on cache or mapping error")
	}

	if len(fields) == 0 {
		return nil, fmt.Errorf("Url not found on cache")
	}

	if err := cmd.Scan(&urlCache); err != nil {
		return nil, fmt.Errorf("Url not found on cache or mapping error")
	}

	return &urlCache, nil
}

func (c *URLCacheRepository) SetUrl(ctx context.Context, url *UrlCache) error {
	exp := time.Duration(config.Env.RedisExpInSeconds) * time.Second

	_, err := c.rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(ctx, url.ShortUrl, url)
		pipe.Expire(ctx, url.ShortUrl, exp)
		return nil
	})

	if err != nil {
		return fmt.Errorf("Error caching url")
	}

	return nil
}
