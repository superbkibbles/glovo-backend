package database

import (
	"github.com/mendmzury/food-delivery/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewPostgres creates a new PostgreSQL connection using GORM
func NewPostgres(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	logger.Info().Str("database", "postgres").Msg("Connected to PostgreSQL")

	return db, nil
}
