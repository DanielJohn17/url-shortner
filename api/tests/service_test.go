package tests

import (
	"context"
	"testing"
	"time"

	"github.com/DanielJohn17/url-shortner/api/internal/urls"
)

func TestServiceCreate(t *testing.T) {
	db := setupTestDB(t)
	svc := urls.NewUrlService(urls.NewURLRepository(db))

	t.Run("creates a 6-char short code", func(t *testing.T) {
		code, err := svc.Create(context.Background(), "https://example.com/x")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(code) != 6 {
			t.Errorf("expected 6-char code, got %q", code)
		}
	})

	t.Run("returns the same code for the same url", func(t *testing.T) {
		a, err := svc.Create(context.Background(), "https://example.com/y")
		if err != nil {
			t.Fatalf("first call failed: %v", err)
		}
		b, err := svc.Create(context.Background(), "https://example.com/y")
		if err != nil {
			t.Fatalf("second call failed: %v", err)
		}
		if a != b {
			t.Errorf("expected identical short codes, got %q and %q", a, b)
		}
	})

	t.Run("canonicalized equivalent urls map to the same code", func(t *testing.T) {
		code, err := svc.Create(context.Background(), "https://example.com/path")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		equiv, err := svc.Create(context.Background(), "https://Example.com:443/path")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != equiv {
			t.Errorf("expected same short code for equivalent urls, got %q and %q", code, equiv)
		}
	})

	t.Run("duplicate create returns the existing short code", func(t *testing.T) {
		svc2 := urls.NewUrlService(urls.NewURLRepository(setupTestDB(t)))

		first, err := svc2.Create(context.Background(), "https://example.com/dup")
		if err != nil {
			t.Fatalf("seed failed: %v", err)
		}
		second, err := svc2.Create(context.Background(), "https://example.com/dup")
		if err != nil {
			t.Fatalf("duplicate create returned an error: %v", err)
		}
		if first != second {
			t.Errorf("expected same short code for duplicate create, got %q and %q", first, second)
		}
	})

	t.Run("accepts empty long url", func(t *testing.T) {
		code, err := svc.Create(context.Background(), "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code == "" {
			t.Error("expected a short code even for empty long url")
		}
	})

	t.Run("rejects malformed url", func(t *testing.T) {
		if _, err := svc.Create(context.Background(), "https://example.com/%zz"); err == nil {
			t.Fatal("expected error for malformed url, got nil")
		}
	})

	t.Run("honors a canceled context", func(t *testing.T) {
		svc2 := urls.NewUrlService(urls.NewURLRepository(setupTestDB(t)))

		if _, err := svc2.Create(context.Background(), "https://example.com/cancel"); err != nil {
			t.Fatalf("seed failed: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		time.Sleep(2 * time.Millisecond)

		if _, err := svc2.Create(ctx, "https://example.com/cancel"); err == nil {
			t.Fatal("expected error with canceled context, got nil")
		}
	})
}

func TestServiceGetUrl(t *testing.T) {
	db := setupTestDB(t)
	svc := urls.NewUrlService(urls.NewURLRepository(db))

	t.Run("returns the long url for an existing short code", func(t *testing.T) {
		code, err := svc.Create(context.Background(), "https://example.com/resolve")
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}

		longUrl, err := svc.GetUrl(context.Background(), code)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if longUrl != "https://example.com/resolve" {
			t.Errorf("got long_url %q", longUrl)
		}
	})

	t.Run("errors on unknown short code", func(t *testing.T) {
		if _, err := svc.GetUrl(context.Background(), "zzzzzz"); err == nil {
			t.Fatal("expected error for unknown short code, got nil")
		}
	})

	t.Run("honors a canceled context", func(t *testing.T) {
		code, err := svc.Create(context.Background(), "https://example.com/resolve-cancel")
		if err != nil {
			t.Fatalf("seed failed: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		time.Sleep(2 * time.Millisecond)

		if _, err := svc.GetUrl(ctx, code); err == nil {
			t.Fatal("expected error with canceled context, got nil")
		}
	})
}
