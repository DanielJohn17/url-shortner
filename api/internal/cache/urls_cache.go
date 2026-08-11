package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrCacheMiss is returned by GetUrl when the short url is not present in the
// cache. It is distinct from real Redis failures so callers can tell a miss
// from a broken cache.
var ErrCacheMiss = errors.New("cache: url not found")

type URLCacheRepositoryInt interface {
	GetUrl(ctx context.Context, shortUrl string) (*UrlCache, error)
	SetUrl(ctx context.Context, url *UrlCache) error
}

type URLCacheRepository struct {
	rdb *redis.Client
	exp time.Duration
}

type UrlCache struct {
	ShortUrl string `redis:"short_url"`
	LongUrl  string `redis:"long_url"`
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
	cmd := c.rdb.HGetAll(ctx, shortUrl)

	fields, err := cmd.Result()
	if err != nil {
		return nil, fmt.Errorf("cache: hgetall %q: %w", shortUrl, err)
	}

	if len(fields) == 0 {
		return nil, ErrCacheMiss
	}

	var urlCache UrlCache
	if err := cmd.Scan(&urlCache); err != nil {
		return nil, fmt.Errorf("cache: scan %q: %w", shortUrl, err)
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
		return fmt.Errorf("cache: set %q: %w", url.ShortUrl, err)
	}

	return nil
}
