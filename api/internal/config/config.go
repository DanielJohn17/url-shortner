package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     string
	DomainName string
}

var Env = initConfig()

func initConfig() Config {
	_ = godotenv.Load("../.env")

	return Config{
		DBHost:     GetEnv("DB_HOST", "localhost"),
		DBUser:     GetEnv("DB_USER", "postgres"),
		DBPassword: GetEnv("DB_PASSWORD", "postgres"),
		DBName:     GetEnv("DB_NAME", "url_shortner_db"),
		DBPort:     GetEnv("DB_PORT", "5432"),
		DomainName: GetEnv("D_Name", "localhost:8080"),
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
