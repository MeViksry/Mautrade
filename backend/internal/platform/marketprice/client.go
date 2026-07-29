package marketprice

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

type Price struct {
	Symbol     string
	PriceQuote string
}

func NewClient() *Client {
	return NewClientWithBaseURL(nil, "https://api.binance.com")
}

func NewClientWithBaseURL(httpClient *http.Client, baseURL string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.binance.com"
	}
	return &Client{httpClient: httpClient, baseURL: baseURL}
}

func (c *Client) FetchSpotPrices(ctx context.Context, symbols []string) ([]Price, error) {
	targets := normalizeSymbols(symbols)
	if len(targets) == 0 {
		return []Price{}, nil
	}

	if len(targets) == 1 {
		return c.fetchSingleSpotPrice(ctx, targets[0])
	}

	prices, err := c.fetchBatchSpotPrices(ctx, targets)
	if err == nil {
		return prices, nil
	}

	var fallback []Price
	var failures []string
	for _, target := range targets {
		price, singleErr := c.fetchSingleSpotPrice(ctx, target)
		if singleErr != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", target.Canonical, singleErr))
			continue
		}
		fallback = append(fallback, price...)
	}
	if len(fallback) > 0 {
		return fallback, nil
	}
	return nil, fmt.Errorf("market price: batch request failed: %v; fallback failed: %s", err, strings.Join(failures, "; "))
}

func (c *Client) fetchBatchSpotPrices(ctx context.Context, targets []symbolTarget) ([]Price, error) {
	binanceSymbols := make([]string, 0, len(targets))
	canonicalByBinance := make(map[string]string, len(targets))
	for _, target := range targets {
		binanceSymbols = append(binanceSymbols, target.Binance)
		canonicalByBinance[target.Binance] = target.Canonical
	}

	symbolsJSON, err := json.Marshal(binanceSymbols)
	if err != nil {
		return nil, fmt.Errorf("market price: encode symbols: %w", err)
	}
	values := url.Values{}
	values.Set("symbols", string(symbolsJSON))
	values.Set("symbolStatus", "TRADING")

	var response []binanceTickerPrice
	if err := c.doJSON(ctx, "/api/v3/ticker/price?"+values.Encode(), &response); err != nil {
		return nil, err
	}

	prices := make([]Price, 0, len(response))
	for _, ticker := range response {
		canonical, ok := canonicalByBinance[strings.ToUpper(strings.TrimSpace(ticker.Symbol))]
		if !ok || strings.TrimSpace(ticker.Price) == "" {
			continue
		}
		prices = append(prices, Price{Symbol: canonical, PriceQuote: strings.TrimSpace(ticker.Price)})
	}
	return prices, nil
}

func (c *Client) fetchSingleSpotPrice(ctx context.Context, target symbolTarget) ([]Price, error) {
	values := url.Values{}
	values.Set("symbol", target.Binance)
	values.Set("symbolStatus", "TRADING")

	var response binanceTickerPrice
	if err := c.doJSON(ctx, "/api/v3/ticker/price?"+values.Encode(), &response); err != nil {
		return nil, err
	}
	if strings.TrimSpace(response.Price) == "" {
		return nil, fmt.Errorf("market price: %s response missing price", target.Canonical)
	}
	return []Price{{Symbol: target.Canonical, PriceQuote: strings.TrimSpace(response.Price)}}, nil
}

func (c *Client) doJSON(ctx context.Context, path string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("market price: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("market price: binance returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("market price: decode response: %w", err)
	}
	return nil
}

type binanceTickerPrice struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

type symbolTarget struct {
	Canonical string
	Binance   string
}

func normalizeSymbols(symbols []string) []symbolTarget {
	seen := map[string]struct{}{}
	targets := make([]symbolTarget, 0, len(symbols))
	for _, symbol := range symbols {
		target, ok := normalizeSymbol(symbol)
		if !ok {
			continue
		}
		if _, exists := seen[target.Canonical]; exists {
			continue
		}
		seen[target.Canonical] = struct{}{}
		targets = append(targets, target)
	}
	sort.Slice(targets, func(i, j int) bool {
		return targets[i].Canonical < targets[j].Canonical
	})
	return targets
}

func normalizeSymbol(symbol string) (symbolTarget, bool) {
	normalized := strings.NewReplacer("-", "/", "_", "/").Replace(strings.ToUpper(strings.TrimSpace(symbol)))
	if normalized == "" {
		return symbolTarget{}, false
	}
	parts := strings.Split(normalized, "/")
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return symbolTarget{Canonical: parts[0] + "/" + parts[1], Binance: parts[0] + parts[1]}, true
	}
	if strings.HasSuffix(normalized, "USDT") && len(normalized) > len("USDT") {
		base := strings.TrimSuffix(normalized, "USDT")
		return symbolTarget{Canonical: base + "/USDT", Binance: base + "USDT"}, true
	}
	return symbolTarget{}, false
}
