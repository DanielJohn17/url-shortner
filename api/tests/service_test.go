package tests

import (
	"context"
	"testing"
	"time"

	"github.com/DanielJohn17/url-shortner/api/internal/urls"
)

func TestServiceCreateWithDelay(t *testing.T) {
	db := setupTestDB(t)
	svc := urls.NewUrlService(urls.NewURLRepository(db))

	t.Run("creates a 6-char short code", func(t *testing.T) {
		code, err := svc.CreateWithDelay(context.Background(), "https://example.com/x")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(code) != 6 {
			t.Errorf("expected 6-char code, got %q", code)
		}
	})

	t.Run("is deterministic for the same url", func(t *testing.T) {
		a, err := svc.CreateWithDelay(context.Background(), "https://example.com/y")
		if err != nil {
			t.Fatalf("first call failed: %v", err)
		}
		b, err := svc.CreateWithDelay(context.Background(), "https://example.com/y")
		if err != nil {
			t.Fatalf("second call failed: %v", err)
		}
		if a != b {
			t.Errorf("expected identical short codes, got %q and %q", a, b)
		}
	})

	t.Run("canonicalized equivalent urls map to the same code", func(t *testing.T) {
		code, err := svc.CreateWithDelay(context.Background(), "https://example.com/path")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		equiv, err := svc.CreateWithDelay(context.Background(), "https://Example.com:443/path")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != equiv {
			t.Errorf("expected same short code for equivalent urls, got %q and %q", code, equiv)
		}
	})

	t.Run("duplicate create returns an error", func(t *testing.T) {
		svc2 := urls.NewUrlService(urls.NewURLRepository(setupTestDB(t)))

		if _, err := svc2.CreateWithDelay(context.Background(), "https://example.com/dup"); err != nil {
			t.Fatalf("seed failed: %v", err)
		}
		if _, err := svc2.CreateWithDelay(context.Background(), "https://example.com/dup"); err == nil {
			t.Fatal("expected error on duplicate create, got nil")
		}
	})

	t.Run("accepts empty long url", func(t *testing.T) {
		code, err := svc.CreateWithDelay(context.Background(), "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code == "" {
			t.Error("expected a short code even for empty long url")
		}
	})

	t.Run("rejects malformed url", func(t *testing.T) {
		if _, err := svc.CreateWithDelay(context.Background(), "https://example.com/%zz"); err == nil {
			t.Fatal("expected error for malformed url, got nil")
		}
	})

	t.Run("honors a canceled context", func(t *testing.T) {
		svc2 := urls.NewUrlService(urls.NewURLRepository(setupTestDB(t)))

		if _, err := svc2.CreateWithDelay(context.Background(), "https://example.com/cancel"); err != nil {
			t.Fatalf("seed failed: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		time.Sleep(2 * time.Millisecond)

		if _, err := svc2.CreateWithDelay(ctx, "https://example.com/cancel"); err == nil {
			t.Fatal("expected error with canceled context, got nil")
		}
	})
}
