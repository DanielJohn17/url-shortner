package tests

import (
	"bytes"
	"context"
	"io"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/DanielJohn17/url-shortner/api/internal/cache"
	"github.com/DanielJohn17/url-shortner/api/internal/router"
	"github.com/DanielJohn17/url-shortner/api/internal/urls"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "github.com/DanielJohn17/url-shortner/api/docs"
)

// testRedisDB is a dedicated Redis DB index used only by the test suite so it
// never collides with real Redis data on the default DB.
const testRedisDB = 15

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

func setupTestRedis(t *testing.T) *redis.Client {
	t.Helper()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Protocol: 3,
		DB:       testRedisDB,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("failed to connect to test redis: %v", err)
	}

	if err := rdb.FlushDB(context.Background()).Err(); err != nil {
		t.Fatalf("failed to flush test redis: %v", err)
	}

	return rdb
}

func newTestURLRepository(t *testing.T) *urls.URLRepository {
	t.Helper()

	db := setupTestDB(t)
	rdb := setupTestRedis(t)
	urlCacheRepo := cache.NewUrlCacheRepository(rdb)

	repo := urls.NewURLRepository(db, urlCacheRepo)
	t.Cleanup(func() { rdb.Close() })

	return repo
}

func buildAppWithDB(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()

	db := setupTestDB(t)
	rdb := setupTestRedis(t)
	urlCacheRepo := cache.NewUrlCacheRepository(rdb)
	t.Cleanup(func() { rdb.Close() })

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
