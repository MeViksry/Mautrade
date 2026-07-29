package store

import "strings"

func exchangeDisplayName(exchange string) string {
	switch strings.ToLower(strings.TrimSpace(exchange)) {
	case "binance":
		return "Binance"
	case "okx":
		return "OKX"
	case "bybit":
		return "Bybit"
	case "tokocrypto":
		return "Tokocrypto"
	default:
		exchange = strings.TrimSpace(exchange)
		if exchange == "" {
			return "Exchange"
		}
		return exchange
	}
}
