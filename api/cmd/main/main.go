package main

import (
	"fmt"

	"github.com/DanielJohn17/url-shortner/api/internal/config"
	"github.com/DanielJohn17/url-shortner/api/internal/router"
	"github.com/DanielJohn17/url-shortner/api/internal/storage"
	"github.com/DanielJohn17/url-shortner/api/internal/urls"
)

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
