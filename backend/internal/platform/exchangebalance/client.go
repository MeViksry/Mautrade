package exchangebalance

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	AccountModeReal    = "real"
	AccountModeDemo    = "demo"
	AccountModeTestnet = "testnet"
)

type BaseURLs struct {
	BinanceReal    string
	BinanceDemo    string
	BinanceTestnet string
	BybitReal      string
	BybitDemo      string
	BybitTestnet   string
	OKX            string
	TokocryptoReal string
}

type Client struct {
	httpClient *http.Client
	baseURLs   BaseURLs
	now        func() time.Time
}

type Credentials struct {
	Exchange    string
	APIKey      string
	APISecret   string
	Passphrase  string
	AccountMode string
	Asset       string
}

type Balance struct {
	Asset       string
	Free        string
	Locked      string
	AccountMode string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURLs: BaseURLs{
			BinanceReal:    "https://api.binance.com",
			BinanceDemo:    "https://testnet.binance.vision",
			BinanceTestnet: "https://testnet.binance.vision",
			BybitReal:      "https://api.bybit.com",
			BybitDemo:      "https://api-demo.bybit.com",
			BybitTestnet:   "https://api-testnet.bybit.com",
			OKX:            "https://openapi.okx.com",
			TokocryptoReal: "https://www.tokocrypto.com",
		},
		now: time.Now,
	}
}

func NewClientWithBaseURLs(httpClient *http.Client, baseURLs BaseURLs) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	client := NewClient()
	client.httpClient = httpClient
	client.baseURLs = mergeBaseURLs(client.baseURLs, baseURLs)
	return client
}

func (c *Client) FetchSpotBalance(ctx context.Context, credentials Credentials) (Balance, error) {
	credentials.Exchange = strings.ToLower(strings.TrimSpace(credentials.Exchange))
	credentials.AccountMode = NormalizeAccountMode(credentials.AccountMode)
	credentials.Asset = strings.ToUpper(strings.TrimSpace(credentials.Asset))
	if credentials.Asset == "" {
		credentials.Asset = "USDT"
	}
	if strings.TrimSpace(credentials.APIKey) == "" || strings.TrimSpace(credentials.APISecret) == "" {
		return Balance{}, fmt.Errorf("exchange balance: api key and secret are required")
	}

	switch credentials.Exchange {
	case "binance":
		return c.fetchBinanceSpotBalance(ctx, credentials)
	case "bybit":
		return c.fetchBybitSpotBalance(ctx, credentials)
	case "okx":
		return c.fetchOKXSpotBalance(ctx, credentials)
	case "tokocrypto":
		return c.fetchTokocryptoSpotBalance(ctx, credentials)
	default:
		return Balance{}, fmt.Errorf("exchange balance: unsupported exchange %q", credentials.Exchange)
	}
}

func NormalizeAccountMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", "real", "live", "production", "prod":
		return AccountModeReal
	case "demo", "paper", "simulated", "simulation":
		return AccountModeDemo
	case "testnet", "sandbox":
		return AccountModeTestnet
	default:
		return strings.ToLower(strings.TrimSpace(mode))
	}
}

func (c *Client) fetchBinanceSpotBalance(ctx context.Context, credentials Credentials) (Balance, error) {
	baseURL := strings.TrimRight(c.baseURLs.BinanceReal, "/")
	resolvedMode := AccountModeReal
	if credentials.AccountMode == AccountModeDemo || credentials.AccountMode == AccountModeTestnet {
		resolvedMode = AccountModeTestnet
		baseURL = strings.TrimRight(firstText(c.baseURLs.BinanceTestnet, c.baseURLs.BinanceDemo), "/")
	}

	values := url.Values{}
	values.Set("omitZeroBalances", "true")
	values.Set("recvWindow", "5000")
	values.Set("timestamp", fmt.Sprintf("%d", c.now().UTC().UnixMilli()))
	queryString := values.Encode()
	signature := hmacSHA256Hex(credentials.APISecret, queryString)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/v3/account?"+queryString+"&signature="+signature, nil)
	if err != nil {
		return Balance{}, err
	}
	req.Header.Set("X-MBX-APIKEY", credentials.APIKey)

	var response binanceAccountResponse
	if err := c.doJSON(req, &response); err != nil {
		return Balance{}, err
	}
	for _, balance := range response.Balances {
		if strings.EqualFold(balance.Asset, credentials.Asset) {
			return balanceWithAccountMode(cleanBalance(credentials.Asset, balance.Free, balance.Locked), resolvedMode), nil
		}
	}
	return balanceWithAccountMode(cleanBalance(credentials.Asset, "0", "0"), resolvedMode), nil
}

func (c *Client) fetchTokocryptoSpotBalance(ctx context.Context, credentials Credentials) (Balance, error) {
	if credentials.AccountMode != AccountModeReal {
		return Balance{}, fmt.Errorf("exchange balance: tokocrypto demo/testnet account is not supported")
	}

	values := url.Values{}
	values.Set("asset", credentials.Asset)
	values.Set("recvWindow", "5000")
	values.Set("timestamp", fmt.Sprintf("%d", c.now().UTC().UnixMilli()))
	queryString := values.Encode()
	signature := hmacSHA256Hex(credentials.APISecret, queryString)
	baseURL := strings.TrimRight(c.baseURLs.TokocryptoReal, "/")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/open/v1/account/spot/asset?"+queryString+"&signature="+signature, nil)
	if err != nil {
		return Balance{}, err
	}
	req.Header.Set("X-MBX-APIKEY", credentials.APIKey)

	var response tokocryptoSpotAssetResponse
	if err := c.doJSON(req, &response); err != nil {
		return Balance{}, err
	}
	if response.Code != 0 {
		return Balance{}, fmt.Errorf("exchange balance: tokocrypto rejected balance request: %s", response.Message)
	}
	if strings.EqualFold(response.Data.Asset, credentials.Asset) {
		return balanceWithAccountMode(cleanBalance(credentials.Asset, response.Data.Free, response.Data.Locked), AccountModeReal), nil
	}
	return balanceWithAccountMode(cleanBalance(credentials.Asset, "0", "0"), AccountModeReal), nil
}

func (c *Client) fetchBybitSpotBalance(ctx context.Context, credentials Credentials) (Balance, error) {
	var zeroBalance *Balance
	var failures []string
	for _, endpoint := range c.bybitBaseURLs(credentials.AccountMode) {
		for _, accountType := range []string{"UNIFIED", "SPOT"} {
			balance, err := c.fetchBybitWalletBalance(ctx, credentials, endpoint.URL, accountType)
			if err != nil {
				failures = append(failures, fmt.Sprintf("%s %s: %v", bybitHostLabel(endpoint.URL), accountType, err))
				continue
			}
			balance = balanceWithAccountMode(balance, endpoint.AccountMode)
			if balance.Free != "0" || balance.Locked != "0" {
				return balance, nil
			}
			if zeroBalance == nil {
				copy := balance
				zeroBalance = &copy
			}
		}
	}
	if zeroBalance != nil {
		return *zeroBalance, nil
	}
	return Balance{}, fmt.Errorf("exchange balance: bybit balance lookup failed: %s", strings.Join(failures, "; "))
}

type bybitEndpoint struct {
	URL         string
	AccountMode string
}

func (c *Client) bybitBaseURLs(accountMode string) []bybitEndpoint {
	endpoints := []bybitEndpoint{{URL: c.baseURLs.BybitReal, AccountMode: AccountModeReal}}
	switch NormalizeAccountMode(accountMode) {
	case AccountModeDemo:
		endpoints = []bybitEndpoint{
			{URL: c.baseURLs.BybitDemo, AccountMode: AccountModeDemo},
			{URL: c.baseURLs.BybitTestnet, AccountMode: AccountModeTestnet},
		}
	case AccountModeTestnet:
		endpoints = []bybitEndpoint{{URL: c.baseURLs.BybitTestnet, AccountMode: AccountModeTestnet}}
	}
	seen := map[string]struct{}{}
	unique := make([]bybitEndpoint, 0, len(endpoints))
	for _, endpoint := range endpoints {
		normalized := strings.TrimRight(strings.TrimSpace(endpoint.URL), "/")
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		unique = append(unique, bybitEndpoint{URL: normalized, AccountMode: endpoint.AccountMode})
	}
	return unique
}

func bybitHostLabel(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return "bybit"
	}
	return parsed.Host
}

func (c *Client) fetchBybitWalletBalance(ctx context.Context, credentials Credentials, baseURL string, accountType string) (Balance, error) {
	baseURL = strings.TrimRight(baseURL, "/")

	values := url.Values{}
	values.Set("accountType", accountType)
	values.Set("coin", credentials.Asset)
	queryString := values.Encode()
	timestamp := fmt.Sprintf("%d", c.now().UTC().UnixMilli())
	recvWindow := "5000"
	signPayload := timestamp + credentials.APIKey + recvWindow + queryString

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/v5/account/wallet-balance?"+queryString, nil)
	if err != nil {
		return Balance{}, err
	}
	req.Header.Set("X-BAPI-API-KEY", credentials.APIKey)
	req.Header.Set("X-BAPI-TIMESTAMP", timestamp)
	req.Header.Set("X-BAPI-RECV-WINDOW", recvWindow)
	req.Header.Set("X-BAPI-SIGN-TYPE", "2")
	req.Header.Set("X-BAPI-SIGN", hmacSHA256Hex(credentials.APISecret, signPayload))

	var response bybitWalletBalanceResponse
	if err := c.doJSON(req, &response); err != nil {
		return Balance{}, err
	}
	if response.RetCode != 0 {
		return Balance{}, fmt.Errorf("exchange balance: bybit rejected balance request: %s", response.RetMsg)
	}
	for _, account := range response.Result.List {
		for _, coin := range account.Coin {
			if strings.EqualFold(coin.Coin, credentials.Asset) {
				free := firstAmount(coin.Free, coin.WalletBalance)
				return cleanBalance(credentials.Asset, free, coin.Locked), nil
			}
		}
	}
	return cleanBalance(credentials.Asset, "0", "0"), nil
}

func (c *Client) fetchOKXSpotBalance(ctx context.Context, credentials Credentials) (Balance, error) {
	if strings.TrimSpace(credentials.Passphrase) == "" {
		return Balance{}, fmt.Errorf("exchange balance: okx passphrase is required")
	}
	resolvedMode := credentials.AccountMode
	if resolvedMode == "" {
		resolvedMode = AccountModeReal
	}
	baseURL := strings.TrimRight(c.baseURLs.OKX, "/")
	pathWithQuery := "/api/v5/account/balance?ccy=" + url.QueryEscape(credentials.Asset)
	timestamp := c.now().UTC().Format("2006-01-02T15:04:05.000Z")
	signature := hmacSHA256Base64(credentials.APISecret, timestamp+http.MethodGet+pathWithQuery)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+pathWithQuery, nil)
	if err != nil {
		return Balance{}, err
	}
	req.Header.Set("OK-ACCESS-KEY", credentials.APIKey)
	req.Header.Set("OK-ACCESS-SIGN", signature)
	req.Header.Set("OK-ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("OK-ACCESS-PASSPHRASE", credentials.Passphrase)
	if credentials.AccountMode == AccountModeDemo || credentials.AccountMode == AccountModeTestnet {
		req.Header.Set("x-simulated-trading", "1")
	}

	var response okxBalanceResponse
	if err := c.doJSON(req, &response); err != nil {
		return Balance{}, err
	}
	if response.Code != "0" {
		return Balance{}, fmt.Errorf("exchange balance: okx rejected balance request: %s", response.Message)
	}
	for _, account := range response.Data {
		for _, detail := range account.Details {
			if strings.EqualFold(detail.Currency, credentials.Asset) {
				free := firstAmount(detail.AvailableBalance, detail.CashBalance)
				locked := firstAmount(detail.FrozenBalance)
				return balanceWithAccountMode(cleanBalance(credentials.Asset, free, locked), resolvedMode), nil
			}
		}
	}
	return balanceWithAccountMode(cleanBalance(credentials.Asset, "0", "0"), resolvedMode), nil
}

func mergeBaseURLs(defaults, overrides BaseURLs) BaseURLs {
	if strings.TrimSpace(overrides.BinanceReal) != "" {
		defaults.BinanceReal = overrides.BinanceReal
	}
	if strings.TrimSpace(overrides.BinanceDemo) != "" {
		defaults.BinanceDemo = overrides.BinanceDemo
	}
	if strings.TrimSpace(overrides.BinanceTestnet) != "" {
		defaults.BinanceTestnet = overrides.BinanceTestnet
	}
	if strings.TrimSpace(overrides.BybitReal) != "" {
		defaults.BybitReal = overrides.BybitReal
	}
	if strings.TrimSpace(overrides.BybitDemo) != "" {
		defaults.BybitDemo = overrides.BybitDemo
	}
	if strings.TrimSpace(overrides.BybitTestnet) != "" {
		defaults.BybitTestnet = overrides.BybitTestnet
	}
	if strings.TrimSpace(overrides.OKX) != "" {
		defaults.OKX = overrides.OKX
	}
	if strings.TrimSpace(overrides.TokocryptoReal) != "" {
		defaults.TokocryptoReal = overrides.TokocryptoReal
	}
	return defaults
}

func (c *Client) doJSON(req *http.Request, target any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("exchange balance: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("exchange balance: %s returned %d: %s", req.URL.Host, resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("exchange balance: decode response: %w", err)
	}
	return nil
}

func cleanBalance(asset, free, locked string) Balance {
	return Balance{
		Asset:  strings.ToUpper(strings.TrimSpace(asset)),
		Free:   firstAmount(free),
		Locked: firstAmount(locked),
	}
}

func balanceWithAccountMode(balance Balance, accountMode string) Balance {
	balance.AccountMode = NormalizeAccountMode(accountMode)
	return balance
}

func firstAmount(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return "0"
}

func firstText(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func hmacSHA256Hex(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func hmacSHA256Base64(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

type binanceAccountResponse struct {
	Balances []struct {
		Asset  string `json:"asset"`
		Free   string `json:"free"`
		Locked string `json:"locked"`
	} `json:"balances"`
}

type tokocryptoSpotAssetResponse struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Data    struct {
		Asset  string `json:"asset"`
		Free   string `json:"free"`
		Locked string `json:"locked"`
	} `json:"data"`
}

type bybitWalletBalanceResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		List []struct {
			AccountType string `json:"accountType"`
			Coin        []struct {
				Coin          string `json:"coin"`
				Free          string `json:"free"`
				Locked        string `json:"locked"`
				WalletBalance string `json:"walletBalance"`
			} `json:"coin"`
		} `json:"list"`
	} `json:"result"`
}

type okxBalanceResponse struct {
	Code    string `json:"code"`
	Message string `json:"msg"`
	Data    []struct {
		Details []struct {
			Currency         string `json:"ccy"`
			AvailableBalance string `json:"availBal"`
			CashBalance      string `json:"cashBal"`
			FrozenBalance    string `json:"frozenBal"`
		} `json:"details"`
	} `json:"data"`
}
