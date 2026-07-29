package marketprice

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestFetchSpotPricesUsesBatchTicker(t *testing.T) {
	t.Parallel()

	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/api/v3/ticker/price" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.URL.Query().Get("symbolStatus") != "TRADING" {
			t.Fatalf("expected trading symbol status")
		}
		if r.URL.Query().Get("symbols") != `["ADAUSDT","ETHUSDT"]` {
			t.Fatalf("unexpected symbols query %q", r.URL.Query().Get("symbols"))
		}
		return jsonResponse(http.StatusOK, []map[string]string{
			{"symbol": "ADAUSDT", "price": "0.16260000"},
			{"symbol": "ETHUSDT", "price": "1889.25000000"},
		})
	})

	client := NewClientWithBaseURL(&http.Client{Transport: transport}, "https://binance.test")
	prices, err := client.FetchSpotPrices(context.Background(), []string{"ETH/USDT", "ada_usdt", "ETHUSDT"})
	if err != nil {
		t.Fatalf("FetchSpotPrices returned error: %v", err)
	}
	if len(prices) != 2 {
		t.Fatalf("expected 2 prices, got %d: %+v", len(prices), prices)
	}
	if prices[0].Symbol != "ADA/USDT" || prices[0].PriceQuote != "0.16260000" {
		t.Fatalf("unexpected first price: %+v", prices[0])
	}
	if prices[1].Symbol != "ETH/USDT" || prices[1].PriceQuote != "1889.25000000" {
		t.Fatalf("unexpected second price: %+v", prices[1])
	}
}

func TestFetchSpotPricesFallsBackToSingleTicker(t *testing.T) {
	t.Parallel()

	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Query().Get("symbols") != "" {
			return textResponse(http.StatusBadRequest, "batch unavailable"), nil
		}
		switch r.URL.Query().Get("symbol") {
		case "ADAUSDT":
			return jsonResponse(http.StatusOK, map[string]string{
				"symbol": "ADAUSDT",
				"price":  "0.16260000",
			})
		case "ETHUSDT":
			return jsonResponse(http.StatusOK, map[string]string{
				"symbol": "ETHUSDT",
				"price":  "1889.25000000",
			})
		default:
			t.Fatalf("unexpected single symbol %q", r.URL.Query().Get("symbol"))
		}
		return textResponse(http.StatusBadRequest, "unexpected symbol"), nil
	})

	client := NewClientWithBaseURL(&http.Client{Transport: transport}, "https://binance.test")
	prices, err := client.FetchSpotPrices(context.Background(), []string{"ETH/USDT", "ADA/USDT"})
	if err != nil {
		t.Fatalf("FetchSpotPrices returned error: %v", err)
	}
	if len(prices) != 2 {
		t.Fatalf("unexpected prices: %+v", prices)
	}
	if prices[0].Symbol != "ADA/USDT" || prices[0].PriceQuote != "0.16260000" {
		t.Fatalf("unexpected first fallback price: %+v", prices[0])
	}
	if prices[1].Symbol != "ETH/USDT" || prices[1].PriceQuote != "1889.25000000" {
		t.Fatalf("unexpected second fallback price: %+v", prices[1])
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func jsonResponse(status int, body any) (*http.Response, error) {
	var builder strings.Builder
	if err := json.NewEncoder(&builder).Encode(body); err != nil {
		return nil, err
	}
	return textResponse(status, builder.String()), nil
}

func textResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}
