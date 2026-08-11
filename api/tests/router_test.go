package tests

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestRouterRoutes(t *testing.T) {
	r := buildApp(t)

	t.Run("POST /api/url_shorter is registered", func(t *testing.T) {
		w := doRequest(t, r, http.MethodPost, "/api/url_shorter", jsonBody(`{"long_url":"https://example.com/r"}`))
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d (body: %s)", w.Code, w.Body.String())
		}
	})

	t.Run("GET /api/url_shorter is not allowed", func(t *testing.T) {
		w := doRequest(t, r, http.MethodGet, "/api/url_shorter", nil)
		if w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 404/405, got %d", w.Code)
		}
	})

	t.Run("GET /api/url_shorter/:shortUrl resolves a stored code", func(t *testing.T) {
		create := doRequest(t, r, http.MethodPost, "/api/url_shorter", jsonBody(`{"long_url":"https://example.com/route"}`))
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
			t.Errorf("expected 302, got %d (body: %s)", w.Code, w.Body.String())
		}
	})

	t.Run("GET /:shortUrl redirects to the long url", func(t *testing.T) {
		create := doRequest(t, r, http.MethodPost, "/api/url_shorter", jsonBody(`{"long_url":"https://example.com/root-redirect"}`))
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

		w := doRequest(t, r, http.MethodGet, "/"+code, nil)
		if w.Code != http.StatusFound {
			t.Fatalf("expected 302, got %d (body: %s)", w.Code, w.Body.String())
		}
		if loc := w.Header().Get("Location"); loc != "https://example.com/root-redirect" {
			t.Errorf("expected Location header %q, got %q", "https://example.com/root-redirect", loc)
		}
	})

	t.Run("GET /swagger/index.html is served", func(t *testing.T) {
		w := doRequest(t, r, http.MethodGet, "/swagger/index.html", nil)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("GET /swagger/doc.json is served", func(t *testing.T) {
		w := doRequest(t, r, http.MethodGet, "/swagger/doc.json", nil)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
		if w.Body.String() == "" {
			t.Error("swagger doc.json should not be empty")
		}
	})

	t.Run("GET /swagger/swagger.yaml is not served by gin-swagger", func(t *testing.T) {
		w := doRequest(t, r, http.MethodGet, "/swagger/swagger.yaml", nil)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("unknown route returns 404", func(t *testing.T) {
		w := doRequest(t, r, http.MethodGet, "/nope", nil)
		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}
