package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Environment            string
	HTTPAddr               string
	DatabaseURL            string
	NATSURL                string
	ShutdownTimeout        time.Duration
	GasFeeShareRate        string
	DefaultCurrency        string
	AllowedCORSOrigin      string
	AuthSessionTTL         time.Duration
	EmailOTPTTL            time.Duration
	GasFeeDepositAddress   string
	ExchangeCredentialKey  string
	AdminBootstrapEmail    string
	AdminBootstrapPassword string
	AdminBootstrapName     string
	AdminBootstrapRole     string
	SMTPHost               string
	SMTPPort               string
	SMTPUsername           string
	SMTPPassword           string
	SMTPFrom               string
	BscScanAPIKey          string
}

func Load() (Config, error) {
	shutdownSeconds, err := intEnv("SHUTDOWN_TIMEOUT_SECONDS", 15)
	if err != nil {
		return Config{}, err
	}
	sessionHours, err := intEnv("AUTH_SESSION_TTL_HOURS", 720)
	if err != nil {
		return Config{}, err
	}
	otpMinutes, err := intEnv("EMAIL_OTP_TTL_MINUTES", 10)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Environment:            stringEnv("APP_ENV", "development"),
		HTTPAddr:               stringEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:            stringEnv("DATABASE_URL", ""),
		NATSURL:                stringEnv("NATS_URL", "nats://localhost:4222"),
		ShutdownTimeout:        time.Duration(shutdownSeconds) * time.Second,
		GasFeeShareRate:        stringEnv("GAS_FEE_SHARE_RATE", "0.5"),
		DefaultCurrency:        stringEnv("DEFAULT_CURRENCY", "USDT"),
		AllowedCORSOrigin:      stringEnv("ALLOWED_CORS_ORIGIN", "*"),
		AuthSessionTTL:         time.Duration(sessionHours) * time.Hour,
		EmailOTPTTL:            time.Duration(otpMinutes) * time.Minute,
		GasFeeDepositAddress:   stringEnv("GAS_FEE_DEPOSIT_ADDRESS", "MAUTRADE-USDT-DEPOSIT-PENDING"),
		ExchangeCredentialKey:  stringEnv("EXCHANGE_CREDENTIAL_KEY", ""),
		AdminBootstrapEmail:    stringEnv("ADMIN_BOOTSTRAP_EMAIL", ""),
		AdminBootstrapPassword: stringEnv("ADMIN_BOOTSTRAP_PASSWORD", ""),
		AdminBootstrapName:     stringEnv("ADMIN_BOOTSTRAP_NAME", "Mautrade Super Admin"),
		AdminBootstrapRole:     stringEnv("ADMIN_BOOTSTRAP_ROLE", "super_admin"),
		SMTPHost:               strings.TrimSpace(stringEnv("SMTP_HOST", "")),
		SMTPPort:               strings.TrimSpace(stringEnv("SMTP_PORT", "587")),
		SMTPUsername:           strings.TrimSpace(stringEnv("SMTP_USERNAME", "")),
		SMTPPassword:           strings.TrimSpace(stringEnv("SMTP_PASSWORD", "")),
		SMTPFrom:               strings.TrimSpace(stringEnv("SMTP_FROM", "verify@mautrade.com")),
		BscScanAPIKey:          strings.TrimSpace(stringEnv("BSCSCAN_API_KEY", "")),
	}, nil
}

func stringEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func intEnv(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("config: %s must be integer: %w", key, err)
	}
	return parsed, nil
}
