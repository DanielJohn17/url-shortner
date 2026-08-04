package router

import (
	"github.com/DanielJohn17/url-shortner/api/internal/urls"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handlers struct {
	URL *urls.URLHandler
}

func NewRoutes(h *Handlers) *gin.Engine {
	router := gin.Default()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	subRouter := router.Group("/api")

	subRouter.POST("/url_shorter", h.URL.Create)

	return router
}
