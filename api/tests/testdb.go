package tests

import (
	"bytes"
	"fmt"
	"io"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DanielJohn17/url-shortner/api/internal/router"
	"github.com/DanielJohn17/url-shortner/api/internal/urls"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "github.com/DanielJohn17/url-shortner/api/docs"
)

var dbCounter int64

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:db_%d_%d?mode=memory&cache=shared", time.Now().UnixNano(), atomic.AddInt64(&dbCounter, 1))

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)

	if err := db.AutoMigrate(&urls.URL{}); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	return db
}

func buildAppWithDB(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()

	db := setupTestDB(t)
	repo := urls.NewURLRepository(db)
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
