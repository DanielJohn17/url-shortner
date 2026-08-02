package router

import (
	"github.com/DanielJohn17/url-shortner/api/internal/urls"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	URL *urls.URLHandler
}

func NewRoutes(h *Handlers) *gin.Engine {
	router := gin.Default()

	subRouter := router.Group("/api")

	subRouter.POST("/url_shorter", h.URL.Create)

	return router
}
