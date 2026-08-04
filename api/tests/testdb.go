package tests

import (
	"bytes"
	"io"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/DanielJohn17/url-shortner/api/internal/router"
	"github.com/DanielJohn17/url-shortner/api/internal/urls"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "github.com/DanielJohn17/url-shortner/api/docs"
)

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
