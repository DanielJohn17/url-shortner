package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type URLCacheRepositoryInt interface {
	GetUrl(ctx context.Context, shortUrl string) (*UrlCache, error)
	SetUrl(ctx context.Context, url *UrlCache) error
}

type URLCacheRepository struct {
	rdb *redis.Client
	exp time.Duration
}

type UrlCache struct {
	ShortUrl  string    `redis:"short_url"`
	LongUrl   string    `redis:"long_url"`
	UsedCount int       `redis:"used_count"`
	CreatedAt time.Time `redis:"created_at"`
}

// asserts *URLCacheRepository satisfies URLCacheRepositoryInt at compile time
var _ URLCacheRepositoryInt = (*URLCacheRepository)(nil)

func NewUrlCacheRepository(rdb *redis.Client, exp time.Duration) URLCacheRepositoryInt {
	return &URLCacheRepository{
		rdb: rdb,
		exp: exp,
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
	_, err := c.rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(ctx, url.ShortUrl, url)
		pipe.Expire(ctx, url.ShortUrl, c.exp)
		return nil
	})

	if err != nil {
		return fmt.Errorf("Error caching url")
	}

	return nil
}
