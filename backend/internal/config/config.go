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
	DefaultCurrency        string
	AllowedCORSOrigin      string
	AuthSessionTTL         time.Duration
	EmailOTPTTL            time.Duration
	GasFeeDepositAddress   string
	GasFeeRPCURLs          []string
	GasFeeChainID          uint64
	GasFeeUSDTContract     string
	GasFeeTokenDecimals    int
	GasFeeMinConfirmations uint64
	GasFeeVerifierInterval time.Duration
	GasFeeWithdrawKey      string
	ExchangeCredentialKey  string
	AdminOneEmail          string
	AdminOneName           string
	AdminOnePassword       string

	AdminTwoEmail    string
	AdminTwoName     string
	AdminTwoPassword string
	SMTPHost         string
	SMTPPort         string
	SMTPUsername     string
	SMTPPassword     string
	SMTPFrom         string
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
	gasFeeChainID, err := uint64Env("GAS_FEE_CHAIN_ID", 56)
	if err != nil {
		return Config{}, err
	}
	gasFeeDecimals, err := intEnv("GAS_FEE_TOKEN_DECIMALS", 18)
	if err != nil {
		return Config{}, err
	}
	gasFeeMinConfirmations, err := uint64Env("GAS_FEE_MIN_CONFIRMATIONS", 5)
	if err != nil {
		return Config{}, err
	}
	gasFeeVerifierSeconds, err := intEnv("GAS_FEE_VERIFIER_INTERVAL_SECONDS", 30)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Environment:            stringEnv("APP_ENV", "development"),
		HTTPAddr:               stringEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:            stringEnv("DATABASE_URL", ""),
		NATSURL:                stringEnv("NATS_URL", "nats://localhost:4222"),
		ShutdownTimeout:        time.Duration(shutdownSeconds) * time.Second,
		DefaultCurrency:        stringEnv("DEFAULT_CURRENCY", "USDT"),
		AllowedCORSOrigin:      stringEnv("ALLOWED_CORS_ORIGIN", "*"),
		AuthSessionTTL:         time.Duration(sessionHours) * time.Hour,
		EmailOTPTTL:            time.Duration(otpMinutes) * time.Minute,
		GasFeeDepositAddress:   stringEnv("GAS_FEE_DEPOSIT_ADDRESS", "MAUTRADE-USDT-DEPOSIT-PENDING"),
		GasFeeRPCURLs:          csvEnv("GAS_FEE_RPC_URLS", []string{"https://bsc-dataseed.bnbchain.org", "https://bsc-dataseed-public.bnbchain.org", "https://bsc-dataseed.nariox.org"}),
		GasFeeChainID:          gasFeeChainID,
		GasFeeUSDTContract:     stringEnv("GAS_FEE_USDT_CONTRACT", "0x55d398326f99059ff775485246999027b3197955"),
		GasFeeTokenDecimals:    gasFeeDecimals,
		GasFeeMinConfirmations: gasFeeMinConfirmations,
		GasFeeVerifierInterval: time.Duration(gasFeeVerifierSeconds) * time.Second,
		GasFeeWithdrawKey:      strings.TrimSpace(stringEnv("GAS_FEE_WITHDRAW_PRIVATE_KEY", "")),
		ExchangeCredentialKey:  stringEnv("EXCHANGE_CREDENTIAL_KEY", ""),
		AdminOneEmail:          stringEnv("ACCOUNT_ADMIN_ONE", ""),
		AdminOneName:           stringEnv("ADMIN_ACCOUNT_ONE_SINGLE_NAME", ""),
		AdminOnePassword:       stringEnv("ADMIN_ACCOUNT_ONE_PASSWORD", ""),

		AdminTwoEmail:    stringEnv("ACCOUNT_ADMIN_TWO", ""),
		AdminTwoName:     stringEnv("ADMIN_ACCOUNT_TWO_SINGLE_NAME", ""),
		AdminTwoPassword: stringEnv("ADMIN_ACCOUNT_TWO_PASSWORD", ""),
		SMTPHost:         strings.TrimSpace(stringEnv("SMTP_HOST", "")),
		SMTPPort:         strings.TrimSpace(stringEnv("SMTP_PORT", "587")),
		SMTPUsername:     strings.TrimSpace(stringEnv("SMTP_USERNAME", "")),
		SMTPPassword:     strings.TrimSpace(stringEnv("SMTP_PASSWORD", "")),
		SMTPFrom:         strings.TrimSpace(stringEnv("SMTP_FROM", "verify@mautrade.com")),
	}, nil
}

func stringEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func csvEnv(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	if len(values) == 0 {
		return fallback
	}
	return values
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

func uint64Env(key string, fallback uint64) (uint64, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("config: %s must be unsigned integer: %w", key, err)
	}
	return parsed, nil
}
