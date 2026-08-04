package main

import (
	"fmt"

	"github.com/DanielJohn17/url-shortner/api/internal/config"
	"github.com/DanielJohn17/url-shortner/api/internal/router"
	"github.com/DanielJohn17/url-shortner/api/internal/storage"
	"github.com/DanielJohn17/url-shortner/api/internal/urls"
	_ "github.com/DanielJohn17/url-shortner/api/docs"
)

// @title URL Shortner API
// @version 1.0
// @description API to shorten long URLs into short codes.
// @host localhost:8080
// @BasePath /api
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

	// auto migrate tables for now
	db.AutoMigrate(&urls.URL{})

	// urls module setup
	urlRepo := urls.NewURLRepository(db)
	urlService := urls.NewUrlService(urlRepo)
	urlHandler := urls.NewUrlHandler(urlService)

	handlers := &router.Handlers{URL: urlHandler}

	r := router.NewRoutes(handlers)

	r.Run(":8080")
}
