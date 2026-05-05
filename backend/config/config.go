package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	Env  string
	DBHost   string
	DBPort   string
	DBUser   string
	DBPass   string
	DBName   string
	DBSSLMode string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	return &Config{
		Port: GetENV("PORT", "8080"),
		Env:  GetENV("ENV", "development"),
		DBHost:    GetENV("DB_HOST", "localhost"),
		DBPort:    GetENV("DB_PORT", "5432"),
		DBUser:    GetENV("DB_USER", "postgres"),
		DBPass:    GetENV("DB_PASS", "postgres"),
		DBName:    GetENV("DB_NAME", "something"),
		DBSSLMode: GetENV("DB_SSLMODE", "disable"),
	}
}

func GetENV(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
