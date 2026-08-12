package router

import (
	"strings"
	"time"

	"github.com/DanielJohn17/url-shortner/api/internal/urls"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handlers struct {
	URL *urls.URLHandler
}

// Config holds the router-level settings sourced from the application config.
type Config struct {
	// AllowedOrigins is a comma-separated list of origins permitted by CORS.
	// Use "*" for open access (local dev) or an explicit list in production
	// (e.g. "https://myapp.vercel.app,https://myapp.com").
	AllowedOrigins string
}

func NewRoutes(h *Handlers, cfg Config) *gin.Engine {
	router := gin.Default()

	// Parse comma-separated origins
	origins := strings.Split(cfg.AllowedOrigins, ",")
	for i, o := range origins {
		origins[i] = strings.TrimSpace(o)
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	subRouter := router.Group("/api")

	subRouter.POST("/url_shorter", h.URL.Create)
	subRouter.GET("/url_shorter/:shortUrl", h.URL.RedirectUrl)

	router.GET("/:shortUrl", h.URL.Redirect)

	return router
}
