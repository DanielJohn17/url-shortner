package main

import (
	"fmt"

	"github.com/DanielJohn17/url-shortner/api/internal/config"
	"github.com/DanielJohn17/url-shortner/api/internal/storage"
	"github.com/DanielJohn17/url-shortner/api/internal/urls"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// benchserver wires the same production postgres-backed handlers and routes as
// the real server, but WITHOUT gin.Default()'s Logger middleware, so we can
// isolate the per-request logging overhead from HTTP+DB throughput.
func main() {
	db, err := storage.NewDatabase(storage.DBConfig{
		DBHost:     config.Env.DBHost,
		DBUser:     config.Env.DBUser,
		DBPassword: config.Env.DBPassword,
		DBName:     config.Env.DBName,
		DBPort:     config.Env.DBPort,
	})
	if err != nil {
		panic(fmt.Sprintf("No database connection: %v", err))
	}

	handler := urls.NewUrlHandler(urls.NewUrlService(urls.NewURLRepository(db)))

	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	sub := router.Group("/api")
	sub.POST("/url_shorter", handler.Create)
	sub.GET("/url_shorter/:shortUrl", handler.RedirectUrl)

	fmt.Println("benchserver on :8081 (no gin Logger middleware)")
	router.Run(":8081")
}