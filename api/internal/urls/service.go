package urls

import (
	"context"
	"fmt"
	"time"

	"github.com/DanielJohn17/url-shortner/api/internal/helpers"
)

type URLServiceInt interface {
	CreateWithDelay(cxt context.Context, longUrl string) (string, error)
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

func (s *URLService) CreateWithDelay(cxt context.Context, longUrl string) (string, error) {
	canonicalUrl, err := helpers.CanonicalizeBasic(longUrl)
	if err != nil {
		return "", err
	}

	maxAttempts := 5
	delay := 1 * time.Millisecond

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		hashCode := helpers.HashString(canonicalUrl)
		shortcode := helpers.EncodeBytes(hashCode[:])

		url := &URL{
			ShortUrl: shortcode[:6],
			LongUrl:  longUrl,
		}

		createdUrl, err := s.repo.Create(cxt, url)
		if err == nil {
			return createdUrl.ShortUrl, err
		}

		fmt.Printf(
			"[Attempt %d/%d] Insert failed: %v. Retrying in %v...\n",
			attempt,
			maxAttempts,
			err,
			delay,
		)

		if attempt < maxAttempts {
			select {
			case <-cxt.Done():
				return "", fmt.Errorf("Failed to create short link")
			case <-time.After(delay):
				delay *= 2
			}
		}

	}

	return "", fmt.Errorf("failed to create short link after %d attempts.", maxAttempts)
}
