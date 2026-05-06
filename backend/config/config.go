package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	Env              string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPass           string
	DBName           string
	DBSSLMode        string
	
	JWTSecret        string
	JWTIssuer        string
	JWTExpiryMinutes string
	RefreshJWTSecret string
	RefreshExpiryHrs string
	RefreshCookie    string


	GoogleClientID	string
	GoogleClientSecret string
	GoogleRedirectURI	string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	return &Config{
		Port:             GetENV("PORT", "8080"),
		Env:              GetENV("ENV", "development"),
		DBHost:           GetENV("DB_HOST", "localhost"),
		DBPort:           GetENV("DB_PORT", "5432"),
		DBUser:           GetENV("DB_USER", "postgres"),
		DBPass:           GetENV("DB_PASS", "postgres"),
		DBName:           GetENV("DB_NAME", "something"),
		DBSSLMode:        GetENV("DB_SSLMODE", "disable"),
		JWTSecret:        GetENV("JWT_SECRET", "dev-super-secret"),
		JWTIssuer:        GetENV("JWT_ISSUER", "something-api"),
		JWTExpiryMinutes: GetENV("JWT_EXPIRY_MINUTES", "15"),
		RefreshJWTSecret: GetENV("REFRESH_JWT_SECRET", "dev-refresh-super-secret"),
		RefreshExpiryHrs: GetENV("REFRESH_EXPIRY_HOURS", "168"),
		RefreshCookie:    GetENV("REFRESH_COOKIE_NAME", "refresh_token"),
		GoogleClientID: GetENV("GOOGLE_CLIENT_ID",""),
		GoogleClientSecret: GetENV("GOOGLE_CLIENT_SECRET",""),
		GoogleRedirectURI:	GetENV("GOOGLE_REDIRECT_URI","http://localhost:8080/auth/google/callback"),
	}
}

func GetENV(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
