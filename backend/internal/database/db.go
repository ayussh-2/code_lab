package database

import (
	"fmt"

	"github.com/ayussh-2/config"
	"github.com/ayussh-2/internal/models"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgres(cfg *config.Config, log *zap.Logger) (*gorm.DB, error) {

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPass,
		cfg.DBName,
		cfg.DBPort,
		cfg.DBSSLMode,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&models.Problems{}); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&models.Topics{}); err != nil {
		return nil, err
	}

	

	if err := db.AutoMigrate(&models.TestCase{}); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&models.OTP{}); err != nil {
		return nil, err
	}

	log.Info("Connected to db and migrated!")
	return db, nil
}
