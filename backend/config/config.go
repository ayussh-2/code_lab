package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

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

	NATSURL          string
	JudgeConcurrency int
	JudgeStream      string
	JudgeSubject     string

	SubmissionRatePerSec     float64
	SubmissionMaxPending     int
	SubmissionSourceMaxBytes int

	SandboxWorkDir          string
	SandboxRunTimeoutMs     int
	SandboxCompileTimeoutMs int
	SandboxMemoryMB         int
	SandboxCPUs             int
	SandboxPidsLimit        int
	SandboxStdoutMaxBytes   int
	SandboxStderrMaxBytes   int
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

		NATSURL:          GetENV("NATS_URL", "nats://localhost:4222"),
		JudgeConcurrency: GetENVInt("JUDGE_CONCURRENCY", 4),
		JudgeStream:      GetENV("JUDGE_STREAM", "JUDGE"),
		JudgeSubject:     GetENV("JUDGE_SUBJECT", "submissions.judge"),

		SubmissionRatePerSec:     GetENVFloat("SUBMISSION_RATE_PER_SEC", 0.34),
		SubmissionMaxPending:     GetENVInt("SUBMISSION_MAX_PENDING", 3),
		SubmissionSourceMaxBytes: GetENVInt("SUBMISSION_SOURCE_MAX_BYTES", 65536),

		SandboxWorkDir:          GetENV("SANDBOX_WORK_DIR", defaultSandboxWorkDir()),
		SandboxRunTimeoutMs:     GetENVInt("SANDBOX_RUN_TIMEOUT_MS", 2000),
		SandboxCompileTimeoutMs: GetENVInt("SANDBOX_COMPILE_TIMEOUT_MS", 10000),
		SandboxMemoryMB:         GetENVInt("SANDBOX_MEMORY_MB", 256),
		SandboxCPUs:             GetENVInt("SANDBOX_CPUS", 1),
		SandboxPidsLimit:        GetENVInt("SANDBOX_PIDS_LIMIT", 128),
		SandboxStdoutMaxBytes:   GetENVInt("SANDBOX_STDOUT_MAX_BYTES", 65536),
		SandboxStderrMaxBytes:   GetENVInt("SANDBOX_STDERR_MAX_BYTES", 65536),
	}
}

func GetENV(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetENVInt(key string, defaultValue int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return defaultValue
	}
	return v
}

func GetENVFloat(key string, defaultValue float64) float64 {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return defaultValue
	}
	return v
}

func defaultSandboxWorkDir() string {
	return filepath.Join(os.TempDir(), "codelab-sandbox")
}
