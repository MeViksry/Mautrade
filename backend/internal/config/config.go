package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Environment       string
	HTTPAddr          string
	DatabaseURL       string
	NATSURL           string
	ShutdownTimeout   time.Duration
	GasFeeShareRate   string
	DefaultCurrency   string
	AllowedCORSOrigin string
}

func Load() (Config, error) {
	shutdownSeconds, err := intEnv("SHUTDOWN_TIMEOUT_SECONDS", 15)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Environment:       stringEnv("APP_ENV", "development"),
		HTTPAddr:          stringEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:       stringEnv("DATABASE_URL", ""),
		NATSURL:           stringEnv("NATS_URL", "nats://localhost:4222"),
		ShutdownTimeout:   time.Duration(shutdownSeconds) * time.Second,
		GasFeeShareRate:   stringEnv("GAS_FEE_SHARE_RATE", "0.5"),
		DefaultCurrency:   stringEnv("DEFAULT_CURRENCY", "USDT"),
		AllowedCORSOrigin: stringEnv("ALLOWED_CORS_ORIGIN", "*"),
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
