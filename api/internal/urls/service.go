package urls

import (
	"context"

	"github.com/DanielJohn17/url-shortner/api/internal/helpers"
)

type URLServiceInt interface {
	Create(cxt context.Context, longUrl string) (string, error)
	GetUrl(cxt context.Context, shortUrl string) (string, error)
}

type URLService struct {
	repo *URLRepository
}

func NewUrlService(r *URLRepository) *URLService {
	return &URLService{
		repo: r,
	}
}

func (s *URLService) Create(cxt context.Context, longUrl string) (string, error) {
	canonicalUrl, err := helpers.CanonicalizeBasic(longUrl)
	if err != nil {
		return "", err
	}

	hashCode := helpers.HashString(canonicalUrl)
	shortcode := helpers.EncodeBytes(hashCode[:])

	existingUrl, err := s.repo.GetUrl(cxt, shortcode[:6])
	if err == nil {
		return existingUrl.ShortUrl, nil
	}

	url := &URL{
		ShortUrl: shortcode[:6],
		LongUrl:  longUrl,
	}

	createdUrl, err := s.repo.Create(cxt, url)
	if err != nil {
		return "", err
	}

	return createdUrl.ShortUrl, nil
}

func (s *URLService) GetUrl(cxt context.Context, shortUrl string) (string, error) {
	url, err := s.repo.GetUrl(cxt, shortUrl)
	if err != nil {
		return "", err
	}

	return url.LongUrl, nil
}
