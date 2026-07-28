package exchangebalance

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchBinanceSpotBalance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/account" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("X-MBX-APIKEY") != "key" {
			t.Fatalf("missing api key header")
		}
		if r.URL.Query().Get("timestamp") != "1234567890000" {
			t.Fatalf("unexpected timestamp")
		}
		if r.URL.Query().Get("signature") == "" {
			t.Fatalf("missing signature")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"balances": []map[string]string{
				{"asset": "BTC", "free": "1", "locked": "0"},
				{"asset": "USDT", "free": "120.50", "locked": "3.25"},
			},
		})
	}))
	defer server.Close()

	client := NewClientWithBaseURLs(server.Client(), BaseURLs{
		BinanceReal: server.URL,
		BinanceDemo: server.URL,
	})
	client.now = func() time.Time { return time.UnixMilli(1234567890000) }

	balance, err := client.FetchSpotBalance(context.Background(), Credentials{
		Exchange:  "binance",
		APIKey:    "key",
		APISecret: "secret",
		Asset:     "USDT",
	})
	if err != nil {
		t.Fatalf("FetchSpotBalance returned error: %v", err)
	}
	if balance.Free != "120.50" || balance.Locked != "3.25" {
		t.Fatalf("unexpected balance: %+v", balance)
	}
}

func TestBinanceDemoUsesDemoModeEndpoint(t *testing.T) {
	client := NewClient()
	if client.baseURLs.BinanceDemo != "https://demo-api.binance.com" {
		t.Fatalf("unexpected binance demo endpoint %s", client.baseURLs.BinanceDemo)
	}
}

func TestBybitDemoUsesMainnetDemoTradingEndpoint(t *testing.T) {
	client := NewClient()
	if client.baseURLs.BybitDemo != "https://api-demo.bybit.com" {
		t.Fatalf("unexpected bybit demo endpoint %s", client.baseURLs.BybitDemo)
	}
	if client.baseURLs.BybitTestnet != "https://api-testnet.bybit.com" {
		t.Fatalf("unexpected bybit testnet endpoint %s", client.baseURLs.BybitTestnet)
	}
}

func TestFetchBybitDemoUnifiedBalance(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/v5/account/wallet-balance" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("X-BAPI-API-KEY") != "key" || r.Header.Get("X-BAPI-SIGN") == "" {
			t.Fatalf("missing bybit signing headers")
		}
		if r.URL.Query().Get("accountType") == "SPOT" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"retCode": 0,
				"retMsg":  "OK",
				"result":  map[string]any{"list": []map[string]any{}},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"retCode": 0,
			"retMsg":  "OK",
			"result": map[string]any{
				"list": []map[string]any{
					{
						"accountType": "UNIFIED",
						"coin": []map[string]string{
							{"coin": "USDT", "walletBalance": "42.75", "locked": "0"},
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewClientWithBaseURLs(server.Client(), BaseURLs{
		BybitReal: server.URL,
		BybitDemo: server.URL,
	})
	client.now = func() time.Time { return time.UnixMilli(1234567890000) }

	balance, err := client.FetchSpotBalance(context.Background(), Credentials{
		Exchange:    "bybit",
		APIKey:      "key",
		APISecret:   "secret",
		AccountMode: "demo",
		Asset:       "USDT",
	})
	if err != nil {
		t.Fatalf("FetchSpotBalance returned error: %v", err)
	}
	if requests != 1 {
		t.Fatalf("expected unified request only, got %d", requests)
	}
	if balance.Free != "42.75" || balance.Locked != "0" {
		t.Fatalf("unexpected balance: %+v", balance)
	}
}

func TestFetchBybitDemoFallsBackToTestnet(t *testing.T) {
	demoRequests := 0
	demoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		demoRequests++
		_ = json.NewEncoder(w).Encode(map[string]any{
			"retCode": 10003,
			"retMsg":  "API key is invalid",
			"result":  map[string]any{"list": []map[string]any{}},
		})
	}))
	defer demoServer.Close()

	testnetRequests := 0
	testnetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testnetRequests++
		if r.URL.Query().Get("accountType") != "UNIFIED" {
			t.Fatalf("expected unified account type first")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"retCode": 0,
			"retMsg":  "OK",
			"result": map[string]any{
				"list": []map[string]any{
					{
						"accountType": "UNIFIED",
						"coin": []map[string]string{
							{"coin": "USDT", "walletBalance": "25.5", "locked": "1.5"},
						},
					},
				},
			},
		})
	}))
	defer testnetServer.Close()

	client := NewClientWithBaseURLs(testnetServer.Client(), BaseURLs{
		BybitDemo:    demoServer.URL,
		BybitTestnet: testnetServer.URL,
	})
	client.now = func() time.Time { return time.UnixMilli(1234567890000) }

	balance, err := client.FetchSpotBalance(context.Background(), Credentials{
		Exchange:    "bybit",
		APIKey:      "key",
		APISecret:   "secret",
		AccountMode: "demo",
		Asset:       "USDT",
	})
	if err != nil {
		t.Fatalf("FetchSpotBalance returned error: %v", err)
	}
	if demoRequests != 2 {
		t.Fatalf("expected demo unified and spot attempts, got %d", demoRequests)
	}
	if testnetRequests != 1 {
		t.Fatalf("expected one testnet request, got %d", testnetRequests)
	}
	if balance.Free != "25.5" || balance.Locked != "1.5" {
		t.Fatalf("unexpected balance: %+v", balance)
	}
}

func TestFetchOKXDemoBalance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v5/account/balance" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("x-simulated-trading") != "1" {
			t.Fatalf("missing okx demo header")
		}
		if r.Header.Get("OK-ACCESS-KEY") != "key" || r.Header.Get("OK-ACCESS-SIGN") == "" || r.Header.Get("OK-ACCESS-PASSPHRASE") != "pass" {
			t.Fatalf("missing okx auth headers")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": "0",
			"msg":  "",
			"data": []map[string]any{
				{
					"details": []map[string]string{
						{"ccy": "USDT", "availBal": "7.5", "frozenBal": "0.25"},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewClientWithBaseURLs(server.Client(), BaseURLs{OKX: server.URL})
	client.now = func() time.Time { return time.UnixMilli(1234567890000) }

	balance, err := client.FetchSpotBalance(context.Background(), Credentials{
		Exchange:    "okx",
		APIKey:      "key",
		APISecret:   "secret",
		Passphrase:  "pass",
		AccountMode: "demo",
		Asset:       "USDT",
	})
	if err != nil {
		t.Fatalf("FetchSpotBalance returned error: %v", err)
	}
	if balance.Free != "7.5" || balance.Locked != "0.25" {
		t.Fatalf("unexpected balance: %+v", balance)
	}
}

func TestTokocryptoDemoUnsupported(t *testing.T) {
	client := NewClient()
	_, err := client.FetchSpotBalance(context.Background(), Credentials{
		Exchange:    "tokocrypto",
		APIKey:      "key",
		APISecret:   "secret",
		AccountMode: "demo",
		Asset:       "USDT",
	})
	if err == nil {
		t.Fatalf("expected tokocrypto demo error")
	}
}
