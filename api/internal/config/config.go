package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost              string
	DBUser              string
	DBPassword          string
	DBName              string
	DBPort              string
	RedisAddr           string
	RedisPassword       string
	RedisDB             int64
	RedisProtocol       int64
	RedisExpInSeconds   int64
	RedisMaxIdleConns   int64
	RedisMaxActiveConns int64
	DomainName          string
}

var Env = initConfig()

func initConfig() Config {
	_ = godotenv.Load("../.env")

	return Config{
		DBHost:              GetEnv("DB_HOST", "localhost"),
		DBUser:              GetEnv("DB_USER", "postgres"),
		DBPassword:          GetEnv("DB_PASSWORD", "postgres"),
		DBName:              GetEnv("DB_NAME", "url_shortner_db"),
		DBPort:              GetEnv("DB_PORT", "5432"),
		RedisAddr:           GetEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:       GetEnv("REDIS_PASS", ""),
		RedisDB:             GetEnvAsInt("REDIS_DB", 0),
		RedisProtocol:       GetEnvAsInt("REDIS_PROTOCOL", 3),
		RedisExpInSeconds:   GetEnvAsInt("REDIS_EXP_IN_SECONDS", 24*3600),
		RedisMaxIdleConns:   GetEnvAsInt("REDIS_MAX_IDLE_CONNS", 10),
		RedisMaxActiveConns: GetEnvAsInt("REDIS_MAX_ACTIVE_CONNS", 100),
		DomainName:          GetEnv("D_Name", "localhost:8080"),
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
