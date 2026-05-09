package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	Env         string
	FrontendURL string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPass      string
	DBName      string
	DBSSLMode   string

	JWTSecret        string
	JWTIssuer        string
	JWTExpiryMinutes string
	RefreshJWTSecret string
	RefreshExpiryHrs string
	RefreshCookie    string
	AccessCookie     string

	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURI  string

	GmailIdentity string
	GmailUserName string
	GmailPassword string
	GmailHost     string
	GmailPort     string
	GmailFrom     string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	return &Config{
		Port:               GetENV("PORT", "8080"),
		Env:                GetENV("ENV", "development"),
		FrontendURL:        GetENV("FRONTEND_URL", "http://localhost:3000"),
		DBHost:             GetENV("DB_HOST", "localhost"),
		DBPort:             GetENV("DB_PORT", "5432"),
		DBUser:             GetENV("DB_USER", "postgres"),
		DBPass:             GetENV("DB_PASS", "postgres"),
		DBName:             GetENV("DB_NAME", "something"),
		DBSSLMode:          GetENV("DB_SSLMODE", "disable"),
		JWTSecret:          GetENV("JWT_SECRET", "dev-super-secret"),
		JWTIssuer:          GetENV("JWT_ISSUER", "something-api"),
		JWTExpiryMinutes:   GetENV("JWT_EXPIRY_MINUTES", "15"),
		RefreshJWTSecret:   GetENV("REFRESH_JWT_SECRET", "dev-refresh-super-secret"),
		RefreshExpiryHrs:   GetENV("REFRESH_EXPIRY_HOURS", "168"),
		RefreshCookie:      GetENV("REFRESH_COOKIE_NAME", "refresh_token"),
		AccessCookie:       GetENV("ACCESS_COOKIE_NAME", "access_token"),
		GoogleClientID:     GetENV("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: GetENV("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURI:  GetENV("GOOGLE_REDIRECT_URI", "http://localhost:8080/auth/google/callback"),

		GmailIdentity: GetENV("GMAIL_IDENTITY", ""),
		GmailUserName: GetENV("GMAIL_EMAIL", ""),
		GmailPassword: GetENV("GMAIL_APP_PASSWORD", ""),
		GmailHost:     GetENV("GMAIL_HOST", "smtp.gmail.com"),
		GmailPort:     GetENV("GMAIL_PORT", "587"),
		GmailFrom:     GetENV("GMAIL_FROM", ""),
	}
}

func GetENV(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
