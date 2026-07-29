package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MeViksry/Mautrade/backend/internal/config"
)

func TestValidateCreateAdminSignalRequestKeepsSellExchangeFilter(t *testing.T) {
	t.Parallel()

	server := &Server{config: config.Config{DefaultCurrency: "USDT"}}
	layerNumber := 1
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/signals", nil)
	params, err := server.validateCreateAdminSignalRequest(request, createAdminSignalRequest{
		Type:           "sell",
		Symbol:         "BTC/USDT",
		LayerNumber:    &layerNumber,
		ExchangeName:   "Binance",
		SellPct:        "100",
		IdempotencyKey: "signal-test",
	}, "00000000-0000-0000-0000-000000000001")
	if err != nil {
		t.Fatalf("expected valid sell signal request, got %v", err)
	}
	if params.ExchangeName != "binance" {
		t.Fatalf("ExchangeName = %q, want binance", params.ExchangeName)
	}
}

func TestValidateCreateAdminSignalRequestRejectsInvalidExchangeFilter(t *testing.T) {
	t.Parallel()

	server := &Server{config: config.Config{DefaultCurrency: "USDT"}}
	layerNumber := 1
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/signals", nil)
	_, err := server.validateCreateAdminSignalRequest(request, createAdminSignalRequest{
		Type:           "sell",
		Symbol:         "BTC/USDT",
		LayerNumber:    &layerNumber,
		ExchangeName:   "unsupported",
		SellPct:        "100",
		IdempotencyKey: "signal-test",
	}, "00000000-0000-0000-0000-000000000001")
	if err == nil {
		t.Fatal("expected invalid exchange filter to fail validation")
	}
}
