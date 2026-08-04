package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/DanielJohn17/url-shortner/api/internal/urls"
)

func TestHandlerCreate(t *testing.T) {
	t.Run("valid url returns 201", func(t *testing.T) {
		r := buildApp(t)
		w := doRequest(t, r, http.MethodPost, "/api/url_shorter", jsonBody(`{"long_url":"https://example.com/x"}`))
		if w.Code != http.StatusCreated {
			t.Fatalf("got status %d, want 201 (body: %s)", w.Code, w.Body.String())
		}

		var body struct {
			Success  bool   `json:"success"`
			ShortURL string `json:"short_url"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("bad json body: %v", err)
		}
		if !body.Success {
			t.Errorf("expected success=true")
		}
		if code, ok := shortCode(t, body.ShortURL); ok && len(code) != 6 {
			t.Errorf("expected 6-char short code, got %q", code)
		}
	})

	t.Run("invalid url format returns 400 and does not persist anything", func(t *testing.T) {
		r, db := buildAppWithDB(t)
		w := doRequest(t, r, http.MethodPost, "/api/url_shorter", jsonBody(`{"long_url":"not-a-url"}`))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want 400 (body: %s)", w.Code, w.Body.String())
		}

		var count int64
		if err := db.Model(&urls.URL{}).Where("long_url = ?", "not-a-url").Count(&count).Error; err != nil {
			t.Fatalf("count failed: %v", err)
		}
		if count != 0 {
			t.Errorf("expected no rows persisted for invalid url, found %d", count)
		}
	})

	t.Run("malformed json returns 400", func(t *testing.T) {
		r := buildApp(t)
		w := doRequest(t, r, http.MethodPost, "/api/url_shorter", jsonBody(`{"long_url":`))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want 400 (body: %s)", w.Code, w.Body.String())
		}
	})

	t.Run("empty body returns 400", func(t *testing.T) {
		r := buildApp(t)
		w := doRequest(t, r, http.MethodPost, "/api/url_shorter", jsonBody(""))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want 400 (body: %s)", w.Code, w.Body.String())
		}
	})

	t.Run("missing long_url field returns 400", func(t *testing.T) {
		r := buildApp(t)
		w := doRequest(t, r, http.MethodPost, "/api/url_shorter", jsonBody(`{}`))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("got status %d, want 400 (body: %s)", w.Code, w.Body.String())
		}
	})

	t.Run("duplicate long url returns the same short code", func(t *testing.T) {
		r, db := buildAppWithDB(t)

		first := doRequest(t, r, http.MethodPost, "/api/url_shorter", jsonBody(`{"long_url":"https://example.com/dup"}`))
		if first.Code != http.StatusCreated {
			t.Fatalf("first create failed: %d %s", first.Code, first.Body.String())
		}

		second := doRequest(t, r, http.MethodPost, "/api/url_shorter", jsonBody(`{"long_url":"https://example.com/dup"}`))
		if second.Code != http.StatusCreated {
			t.Fatalf("expected 201 on duplicate, got %d (body: %s)", second.Code, second.Body.String())
		}

		var firstBody, secondBody struct {
			ShortURL string `json:"short_url"`
		}
		if err := json.Unmarshal(first.Body.Bytes(), &firstBody); err != nil {
			t.Fatalf("bad first body: %v", err)
		}
		if err := json.Unmarshal(second.Body.Bytes(), &secondBody); err != nil {
			t.Fatalf("bad second body: %v", err)
		}
		if firstBody.ShortURL != secondBody.ShortURL {
			t.Errorf("expected same short code on duplicate, got %q and %q", firstBody.ShortURL, secondBody.ShortURL)
		}

		var count int64
		if err := db.Model(&urls.URL{}).Where("long_url = ?", "https://example.com/dup").Count(&count).Error; err != nil {
			t.Fatalf("count failed: %v", err)
		}
		if count != 1 {
			t.Errorf("expected a single stored row for one long url, found %d (duplicate rows stored)", count)
		}
	})

	t.Run("canonicalized equivalent url deduplicates", func(t *testing.T) {
		r, db := buildAppWithDB(t)

		var first, second struct {
			ShortURL string `json:"short_url"`
		}

		if w := doRequest(t, r, http.MethodPost, "/api/url_shorter", jsonBody(`{"long_url":"https://example.com/canon"}`)); w.Code != http.StatusCreated {
			t.Fatalf("first create failed: %d %s", w.Code, w.Body.String())
		} else if err := json.Unmarshal(w.Body.Bytes(), &first); err != nil {
			t.Fatalf("bad first body: %v", err)
		}
		if w := doRequest(t, r, http.MethodPost, "/api/url_shorter", jsonBody(`{"long_url":"https://EXAMPLE.com:443/canon"}`)); w.Code != http.StatusCreated {
			t.Fatalf("second create failed: %d %s", w.Code, w.Body.String())
		} else if err := json.Unmarshal(w.Body.Bytes(), &second); err != nil {
			t.Fatalf("bad second body: %v", err)
		}

		if first.ShortURL != second.ShortURL {
			t.Errorf("expected same short code for equivalent urls, got %q and %q", first.ShortURL, second.ShortURL)
		}

		var count int64
		if err := db.Model(&urls.URL{}).Where("long_url IN ?", []string{"https://example.com/canon", "https://EXAMPLE.com:443/canon"}).Count(&count).Error; err != nil {
			t.Fatalf("count failed: %v", err)
		}
		if count != 1 {
			t.Errorf("expected 1 stored row for canonical-equivalent urls, got %d", count)
		}
	})
}

func TestHandlerStoredShortUrlIsResolvable(t *testing.T) {
	r, db := buildAppWithDB(t)

	w := doRequest(t, r, http.MethodPost, "/api/url_shorter", jsonBody(`{"long_url":"https://example.com/resolve"}`))
	if w.Code != http.StatusCreated {
		t.Fatalf("create failed: %d %s", w.Code, w.Body.String())
	}

	var body struct {
		ShortURL string `json:"short_url"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}

	code, ok := shortCode(t, body.ShortURL)
	if !ok {
		t.Fatal("could not extract short code")
	}

	repo := urls.NewURLRepository(db)
	got, err := repo.GetUrl(context.Background(), code)
	if err != nil {
		t.Fatalf("stored short url not resolvable: %v", err)
	}
	if got.LongUrl != "https://example.com/resolve" {
		t.Errorf("got long_url %q", got.LongUrl)
	}
}

func TestHandlerRedirectUrl(t *testing.T) {
	r, _ := buildAppWithDB(t)

	t.Run("returns 302 with the long url", func(t *testing.T) {
		create := doRequest(t, r, http.MethodPost, "/api/url_shorter", jsonBody(`{"long_url":"https://example.com/redirect"}`))
		if create.Code != http.StatusCreated {
			t.Fatalf("create failed: %d %s", create.Code, create.Body.String())
		}

		var created struct {
			ShortURL string `json:"short_url"`
		}
		if err := json.Unmarshal(create.Body.Bytes(), &created); err != nil {
			t.Fatal(err)
		}
		code, ok := shortCode(t, created.ShortURL)
		if !ok {
			t.Fatal("could not extract short code")
		}

		w := doRequest(t, r, http.MethodGet, "/api/url_shorter/"+code, nil)
		if w.Code != http.StatusFound {
			t.Fatalf("got status %d, want 302 (body: %s)", w.Code, w.Body.String())
		}

		var body struct {
			Success bool   `json:"success"`
			LongURL string `json:"long_url"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if !body.Success {
			t.Errorf("expected success=true")
		}
		if body.LongURL != "https://example.com/redirect" {
			t.Errorf("got long_url %q", body.LongURL)
		}
	})

	t.Run("unknown short code returns 404", func(t *testing.T) {
		w := doRequest(t, r, http.MethodGet, "/api/url_shorter/zzzzzz", nil)
		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want 404 (body: %s)", w.Code, w.Body.String())
		}
	})
}
