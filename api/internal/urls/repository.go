package urls

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type URLRepositoryInt interface {
	Create(ctx context.Context, url *URL) (*URL, error)
	GetUrl(cxt context.Context, shortUrl string) (*URL, error)
}

type URLRepository struct {
	db *gorm.DB
}

func NewURLRepository(db *gorm.DB) *URLRepository {
	return &URLRepository{
		db: db,
	}
}

func (r *URLRepository) Create(cxt context.Context, url *URL) (*URL, error) {

	err := gorm.G[URL](r.db).Create(cxt, url)
	if err != nil {
		return nil, fmt.Errorf("Failed to create short link")
	}

	return url, nil

}

func (r *URLRepository) GetUrl(cxt context.Context, shortUrl string) (*URL, error) {
	url, err := gorm.G[URL](r.db).Where("short_url = ?", shortUrl).First(cxt)
	if err != nil {
		return nil, fmt.Errorf("Redirect not found")
	}

	return &url, nil
}
