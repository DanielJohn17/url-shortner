package tests

import (
	"context"
	"testing"

	"github.com/DanielJohn17/url-shortner/api/internal/urls"
)

func TestRepositoryCreate(t *testing.T) {
	t.Run("creates new url", func(t *testing.T) {
		repo := newTestURLRepository(t)

		created, err := repo.Create(context.Background(), &urls.URL{ShortUrl: "abc123", LongUrl: "https://example.com/x"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if created.ShortUrl != "abc123" {
			t.Errorf("got short_url %q", created.ShortUrl)
		}
	})

	t.Run("duplicate short url returns error", func(t *testing.T) {
		repo := newTestURLRepository(t)
		ctx := context.Background()

		if _, err := repo.Create(ctx, &urls.URL{ShortUrl: "abc123", LongUrl: "https://example.com/x"}); err != nil {
			t.Fatalf("seed failed: %v", err)
		}

		if _, err := repo.Create(ctx, &urls.URL{ShortUrl: "abc123", LongUrl: "https://example.com/y"}); err == nil {
			t.Fatal("expected duplicate-key error, got nil")
		}
	})

	t.Run("empty short url", func(t *testing.T) {
		repo := newTestURLRepository(t)

		created, err := repo.Create(context.Background(), &urls.URL{ShortUrl: "", LongUrl: "https://example.com/z"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if created.ShortUrl != "" {
			t.Errorf("got %q", created.ShortUrl)
		}
	})
}

func TestRepositoryGetUrl(t *testing.T) {
	t.Run("finds existing url", func(t *testing.T) {
		repo := newTestURLRepository(t)
		ctx := context.Background()

		if _, err := repo.Create(ctx, &urls.URL{ShortUrl: "findme", LongUrl: "https://example.com/find"}); err != nil {
			t.Fatalf("seed failed: %v", err)
		}

		got, err := repo.GetUrl(ctx, "findme")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.LongUrl != "https://example.com/find" {
			t.Errorf("got long_url %q", got.LongUrl)
		}
	})

	t.Run("missing url returns error", func(t *testing.T) {
		repo := newTestURLRepository(t)

		if _, err := repo.GetUrl(context.Background(), "missing"); err == nil {
			t.Fatal("expected not-found error, got nil")
		}
	})

	t.Run("empty short url returns error", func(t *testing.T) {
		repo := newTestURLRepository(t)

		if _, err := repo.GetUrl(context.Background(), ""); err == nil {
			t.Fatal("expected error for empty short url, got nil")
		}
	})
}
