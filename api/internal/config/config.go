package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Postgres — individual fields used for local dev.
	// Set DATABASE_URL to use a connection string (e.g. Neon) instead.
	DatabaseURL string
	DBHost      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBPort      string

	// Redis — individual fields used for local dev.
	// Set REDIS_URL to use a connection string (e.g. Upstash TCP) instead.
	RedisURL            string
	RedisAddr           string
	RedisPassword       string
	RedisDB             int64
	RedisProtocol       int64
	RedisExpInSeconds   int64
	RedisMaxIdleConns   int64
	RedisMaxActiveConns int64

	DomainName string
	Port        string
}

var Env = initConfig()

func initConfig() Config {
	_ = godotenv.Load(".env")

	return Config{
		// Postgres
		DatabaseURL: GetEnv("DATABASE_URL", ""),
		DBHost:      GetEnv("DB_HOST", "localhost"),
		DBUser:      GetEnv("DB_USER", "postgres"),
		DBPassword:  GetEnv("DB_PASSWORD", "postgres"),
		DBName:      GetEnv("DB_NAME", "url_shortner_db"),
		DBPort:      GetEnv("DB_PORT", "5432"),

		// Redis
		RedisURL:            GetEnv("REDIS_URL", ""),
		RedisAddr:           GetEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:       GetEnv("REDIS_PASS", ""),
		RedisDB:             GetEnvAsInt("REDIS_DB", 0),
		RedisProtocol:       GetEnvAsInt("REDIS_PROTOCOL", 3),
		RedisExpInSeconds:   GetEnvAsInt("REDIS_EXP_IN_SECONDS", 86400),
		RedisMaxIdleConns:   GetEnvAsInt("REDIS_MAX_IDLE_CONNS", 10),
		RedisMaxActiveConns: GetEnvAsInt("REDIS_MAX_ACTIVE_CONNS", 100),

		DomainName: GetEnv("DOMAIN_NAME", "localhost:8080"),
		Port:       GetEnv("PORT", "8080"),
	}
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		if len(value) == 0 {
			return fallback
		}
		return value
	}

	return fallback
}

func GetEnvAsInt(key string, fallback int64) int64 {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	valInt, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}

	return valInt
}
