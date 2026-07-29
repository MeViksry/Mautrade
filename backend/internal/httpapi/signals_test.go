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

func TestValidateCreateAdminSignalRequestIgnoresBuyLayerNumber(t *testing.T) {
	t.Parallel()

	server := &Server{config: config.Config{DefaultCurrency: "USDT"}}
	layerNumber := 999
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/signals", nil)
	params, err := server.validateCreateAdminSignalRequest(request, createAdminSignalRequest{
		Type:           "buy",
		Symbol:         "BTC/USDT",
		LayerNumber:    &layerNumber,
		AllocationPct:  "10",
		IdempotencyKey: "signal-test",
	}, "00000000-0000-0000-0000-000000000001")
	if err != nil {
		t.Fatalf("expected valid buy signal request, got %v", err)
	}
	if params.LayerNumber != nil {
		t.Fatalf("buy LayerNumber = %v, want nil", *params.LayerNumber)
	}
}

func TestExecutionResultBalanceAssetsIncludesQuoteAndBase(t *testing.T) {
	t.Parallel()

	assets := executionResultBalanceAssets("USDT", "BTC/USDT")
	assertStringSlice(t, assets, []string{"USDT", "BTC"})
}

func TestExecutionResultBalanceAssetsAvoidsDuplicateDefaultAsset(t *testing.T) {
	t.Parallel()

	assets := executionResultBalanceAssets("USDT", "USDT/IDR")
	assertStringSlice(t, assets, []string{"USDT"})
}

func TestBaseAssetFromSignalSymbolAcceptsExchangeSeparators(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"BTC/USDT": "BTC",
		"eth-usdt": "ETH",
		"sol_usdt": "SOL",
	}
	for symbol, expected := range cases {
		if got := baseAssetFromSignalSymbol(symbol); got != expected {
			t.Fatalf("baseAssetFromSignalSymbol(%q) = %q, want %q", symbol, got, expected)
		}
	}
}

func assertStringSlice(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
