package storage

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBConfig struct {
	// DatabaseURL is an optional full connection string (e.g. Neon postgres://...).
	// When non-empty it takes priority and all other fields are ignored.
	DatabaseURL string
	DBHost      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBPort      string
}

func NewDatabase(config DBConfig) (*gorm.DB, error) {
	var dsn string

	if config.DatabaseURL != "" {
		// Neon / any Postgres URL — SSL/TLS params are already embedded in the URL
		dsn = config.DatabaseURL
	} else {
		// Local Postgres via individual fields
		dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			config.DBHost,
			config.DBUser,
			config.DBPassword,
			config.DBName,
			config.DBPort,
		)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(100)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	return db, nil
}
