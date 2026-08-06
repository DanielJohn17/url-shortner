package urls

import (
	"context"
	"fmt"

	"github.com/DanielJohn17/url-shortner/api/internal/cache"
	"gorm.io/gorm"
)

type URLRepositoryInt interface {
	Create(ctx context.Context, url *URL) (*URL, error)
	GetUrl(ctx context.Context, shortUrl string) (*URL, error)
}

type URLRepository struct {
	db    *gorm.DB
	cache *cache.URLCacheRepository
}

func NewURLRepository(db *gorm.DB, cache *cache.URLCacheRepository) *URLRepository {
	return &URLRepository{
		db:    db,
		cache: cache,
	}
}

func (r *URLRepository) Create(ctx context.Context, url *URL) (*URL, error) {

	err := gorm.G[URL](r.db).Create(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("Failed to create short link")
	}

	r.saveToCache(ctx, url)

	return url, nil

}

func (r *URLRepository) GetUrl(ctx context.Context, shortUrl string) (*URL, error) {

	urlCache, err := r.cache.GetUrl(ctx, shortUrl)
	if err == nil {
		return &URL{
			ShortUrl:  urlCache.ShortUrl,
			LongUrl:   urlCache.LongUrl,
			UsedCount: urlCache.UsedCount,
			CreatedAt: urlCache.CreatedAt,
		}, nil
	}

	url, err := gorm.G[URL](r.db).Where("short_url = ?", shortUrl).First(ctx)
	if err != nil {
		return nil, fmt.Errorf("Redirect url not found")
	}

	r.saveToCache(ctx, &url)

	return &url, nil
}

func (r *URLRepository) saveToCache(ctx context.Context, url *URL) {
	urlCache := &cache.UrlCache{
		ShortUrl:  url.ShortUrl,
		LongUrl:   url.LongUrl,
		UsedCount: url.UsedCount,
		CreatedAt: url.CreatedAt,
	}

	if err := r.cache.SetUrl(ctx, urlCache); err != nil {
		fmt.Println("!!!!Error Saving URL to Cache!!!!")
	}
}
