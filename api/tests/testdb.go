package tests

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"

	"github.com/DanielJohn17/url-shortner/api/internal/cache"
	"github.com/DanielJohn17/url-shortner/api/internal/router"
	"github.com/DanielJohn17/url-shortner/api/internal/urls"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "github.com/DanielJohn17/url-shortner/api/docs"
)

// fakeURLCache is an in-memory implementation of cache.URLCacheRepositoryInt
// used so tests run without a live Redis server.
type fakeURLCache struct {
	mu    sync.Mutex
	items map[string]cache.UrlCache
}

var _ cache.URLCacheRepositoryInt = (*fakeURLCache)(nil)

func newFakeURLCache() *fakeURLCache {
	return &fakeURLCache{items: make(map[string]cache.UrlCache)}
}

func (f *fakeURLCache) GetUrl(ctx context.Context, shortUrl string) (*cache.UrlCache, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	item, ok := f.items[shortUrl]
	if !ok {
		return nil, fmt.Errorf("Url not found on cache")
	}

	return &item, nil
}

func (f *fakeURLCache) SetUrl(ctx context.Context, url *cache.UrlCache) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.items[url.ShortUrl] = *url
	return nil
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	if err := db.AutoMigrate(&urls.URL{}); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	return db
}

func newTestURLRepository(t *testing.T) *urls.URLRepository {
	t.Helper()

	db := setupTestDB(t)
	urlCacheRepo := newFakeURLCache()

	return urls.NewURLRepository(db, urlCacheRepo)
}

func buildAppWithDB(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()

	db := setupTestDB(t)
	urlCacheRepo := newFakeURLCache()

	repo := urls.NewURLRepository(db, urlCacheRepo)
	svc := urls.NewUrlService(repo)
	h := urls.NewUrlHandler(svc)

	return router.NewRoutes(&router.Handlers{URL: h}), db
}

func buildApp(t *testing.T) *gin.Engine {
	t.Helper()

	r, _ := buildAppWithDB(t)
	return r
}

func doRequest(t *testing.T, r *gin.Engine, method, path string, body io.Reader) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

func jsonBody(s string) io.Reader {
	return bytes.NewBufferString(s)
}
