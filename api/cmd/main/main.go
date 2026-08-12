package main

import (
	"fmt"
	"time"

	_ "github.com/DanielJohn17/url-shortner/api/docs"
	"github.com/DanielJohn17/url-shortner/api/internal/cache"
	"github.com/DanielJohn17/url-shortner/api/internal/config"
	"github.com/DanielJohn17/url-shortner/api/internal/router"
	"github.com/DanielJohn17/url-shortner/api/internal/storage"
	"github.com/DanielJohn17/url-shortner/api/internal/urls"
)

// @title URL Shortner API
// @version 1.0
// @description API to shorten long URLs into short codes.
// @host ${DOMAIN_NAME}
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

	rdb, err := cache.NewCacheStorage(cache.RedisConfig{
		Addr:           config.Env.RedisAddr,
		Password:       config.Env.RedisPassword,
		DB:             int(config.Env.RedisDB),
		Protocol:       int(config.Env.RedisProtocol),
		MaxIdleConns:   int(config.Env.RedisMaxIdleConns),
		MaxActiveConns: int(config.Env.RedisMaxActiveConns),
	})

	if err != nil {
		panic(fmt.Sprintf("Error connecting to redis: %v", err))
	}

	defer func() {
		if err := rdb.Close(); err != nil {
			fmt.Printf("Error closing redis connection: %v\n", err)
		}
	}()

	// auto migrate tables for now
	db.AutoMigrate(&urls.URL{})

	// cache module setup
	urlCacheRepo := cache.NewUrlCacheRepository(rdb, time.Duration(config.Env.RedisExpInSeconds)*time.Second)

	// urls module setup
	urlRepo := urls.NewURLRepository(db, urlCacheRepo)
	urlService := urls.NewUrlService(urlRepo)
	urlHandler := urls.NewUrlHandler(urlService)

	handlers := &router.Handlers{URL: urlHandler}

	r := router.NewRoutes(handlers)

	r.Run(":" + config.Env.Port)
}
