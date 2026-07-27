package bscrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidTransaction = errors.New("bscrpc: invalid transaction")
	ErrNetwork            = errors.New("bscrpc: network or rpc error")
	ErrPending            = errors.New("bscrpc: transaction pending")
)

const (
	DefaultUSDTContractAddress = "0x55d398326f99059ff775485246999027b3197955"
	TransferEventTopic         = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
)

type Config struct {
	RPCURLs          []string
	ChainID          uint64
	TokenContract    string
	TokenDecimals    int
	MinConfirmations uint64
	HTTPClient       *http.Client
}

type Client struct {
	httpClient *http.Client
	rpcURLs    []string
	chainID    uint64
	token      string
	minConf    uint64
}

type TransferVerification struct {
	TxID          string
	ChainID       uint64
	TokenContract string
	FromAddress   string
	ToAddress     string
	Amount        *big.Int
	BlockNumber   uint64
	Confirmations uint64
}

type jsonrpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
	ID      int    `json:"id"`
}

type jsonrpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type transactionReceipt struct {
	TransactionHash string       `json:"transactionHash"`
	Status          string       `json:"status"`
	BlockNumber     string       `json:"blockNumber"`
	From            string       `json:"from"`
	To              string       `json:"to"`
	Logs            []receiptLog `json:"logs"`
}

type receiptLog struct {
	Address string   `json:"address"`
	Topics  []string `json:"topics"`
	Data    string   `json:"data"`
}

func NewClient(config Config) *Client {
	rpcURLs := cleanRPCURLs(config.RPCURLs)
	if len(rpcURLs) == 0 {
		rpcURLs = []string{"https://bsc-dataseed.bnbchain.org"}
	}
	chainID := config.ChainID
	if chainID == 0 {
		chainID = 56
	}
	token := strings.TrimSpace(config.TokenContract)
	if token == "" {
		token = DefaultUSDTContractAddress
	}
	minConf := config.MinConfirmations
	if minConf == 0 {
		minConf = 1
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{
		httpClient: httpClient,
		rpcURLs:    rpcURLs,
		chainID:    chainID,
		token:      token,
		minConf:    minConf,
	}
}

func (c *Client) VerifyUSDTTransfer(ctx context.Context, txID, expectedRecipient string) (TransferVerification, error) {
	normalizedTxID, err := NormalizeTxHash(txID)
	if err != nil {
		return TransferVerification{}, err
	}
	recipient, err := normalizeAddress(expectedRecipient)
	if err != nil {
		return TransferVerification{}, fmt.Errorf("%w: invalid recipient address: %v", ErrInvalidTransaction, err)
	}
	token, err := normalizeAddress(c.token)
	if err != nil {
		return TransferVerification{}, fmt.Errorf("%w: invalid token contract: %v", ErrInvalidTransaction, err)
	}

	chainID, err := c.chainIDFromRPC(ctx)
	if err != nil {
		return TransferVerification{}, err
	}
	if chainID != c.chainID {
		return TransferVerification{}, fmt.Errorf("%w: chain id mismatch got %d expected %d", ErrNetwork, chainID, c.chainID)
	}

	receipt, err := c.transactionReceipt(ctx, normalizedTxID)
	if err != nil {
		return TransferVerification{}, err
	}
	if receipt == nil {
		return TransferVerification{}, fmt.Errorf("%w: transaction receipt not found", ErrPending)
	}
	if strings.EqualFold(receipt.Status, "0x0") {
		return TransferVerification{}, fmt.Errorf("%w: transaction reverted on-chain", ErrInvalidTransaction)
	}
	if !strings.EqualFold(receipt.Status, "0x1") {
		return TransferVerification{}, fmt.Errorf("%w: transaction status %s", ErrPending, receipt.Status)
	}

	blockNumber, err := parseHexUint64(receipt.BlockNumber)
	if err != nil {
		return TransferVerification{}, fmt.Errorf("%w: invalid receipt block number: %v", ErrNetwork, err)
	}
	latestBlock, err := c.blockNumber(ctx)
	if err != nil {
		return TransferVerification{}, err
	}
	confirmations := uint64(0)
	if latestBlock >= blockNumber {
		confirmations = latestBlock - blockNumber + 1
	}
	if confirmations < c.minConf {
		return TransferVerification{}, fmt.Errorf("%w: confirmations %d/%d", ErrPending, confirmations, c.minConf)
	}

	totalAmount := new(big.Int)
	firstSender := ""
	for _, log := range receipt.Logs {
		if !strings.EqualFold(log.Address, token) || len(log.Topics) < 3 {
			continue
		}
		if !strings.EqualFold(log.Topics[0], TransferEventTopic) {
			continue
		}
		toAddress, err := addressFromTopic(log.Topics[2])
		if err != nil || !strings.EqualFold(toAddress, recipient) {
			continue
		}
		amount, err := uint256FromHex(log.Data)
		if err != nil {
			return TransferVerification{}, fmt.Errorf("%w: invalid transfer amount: %v", ErrInvalidTransaction, err)
		}
		if firstSender == "" {
			if fromAddress, err := addressFromTopic(log.Topics[1]); err == nil {
				firstSender = fromAddress
			}
		}
		totalAmount = totalAmount.Add(totalAmount, amount)
	}
	if totalAmount.Sign() <= 0 {
		return TransferVerification{}, fmt.Errorf("%w: no USDT BEP-20 transfer to expected recipient", ErrInvalidTransaction)
	}

	return TransferVerification{
		TxID:          normalizedTxID,
		ChainID:       chainID,
		TokenContract: token,
		FromAddress:   firstSender,
		ToAddress:     recipient,
		Amount:        totalAmount,
		BlockNumber:   blockNumber,
		Confirmations: confirmations,
	}, nil
}

func NormalizeTxHash(value string) (string, error) {
	hash := strings.ToLower(strings.TrimSpace(value))
	if hash == "" {
		return "", fmt.Errorf("%w: tx_id is required", ErrInvalidTransaction)
	}
	if !strings.HasPrefix(hash, "0x") {
		hash = "0x" + hash
	}
	if len(hash) != 66 {
		return "", fmt.Errorf("%w: tx_id must be 32-byte hex hash", ErrInvalidTransaction)
	}
	for _, r := range hash[2:] {
		if !isHex(r) {
			return "", fmt.Errorf("%w: tx_id must be hex", ErrInvalidTransaction)
		}
	}
	return hash, nil
}

func (c *Client) chainIDFromRPC(ctx context.Context) (uint64, error) {
	var value string
	if err := c.rpc(ctx, "eth_chainId", nil, &value); err != nil {
		return 0, err
	}
	parsed, err := parseHexUint64(value)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid chain id: %v", ErrNetwork, err)
	}
	return parsed, nil
}

func (c *Client) transactionReceipt(ctx context.Context, txID string) (*transactionReceipt, error) {
	var receipt *transactionReceipt
	if err := c.rpc(ctx, "eth_getTransactionReceipt", []any{txID}, &receipt); err != nil {
		return nil, err
	}
	return receipt, nil
}

func (c *Client) blockNumber(ctx context.Context) (uint64, error) {
	var value string
	if err := c.rpc(ctx, "eth_blockNumber", nil, &value); err != nil {
		return 0, err
	}
	parsed, err := parseHexUint64(value)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid block number: %v", ErrNetwork, err)
	}
	return parsed, nil
}

func (c *Client) rpc(ctx context.Context, method string, params []any, target any) error {
	var lastErr error
	for _, endpoint := range c.rpcURLs {
		if err := c.rpcOne(ctx, endpoint, method, params, target); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	if lastErr == nil {
		lastErr = errors.New("no rpc endpoints configured")
	}
	return fmt.Errorf("%w: %v", ErrNetwork, lastErr)
}

func (c *Client) rpcOne(ctx context.Context, endpoint, method string, params []any, target any) error {
	if params == nil {
		params = []any{}
	}
	body, err := json.Marshal(jsonrpcRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("rpc http status %d", res.StatusCode)
	}
	var envelope jsonrpcResponse
	if err := json.NewDecoder(res.Body).Decode(&envelope); err != nil {
		return fmt.Errorf("decode rpc response: %w", err)
	}
	if envelope.Error != nil {
		return fmt.Errorf("rpc error: %s", envelope.Error.Message)
	}
	if len(envelope.Result) == 0 {
		return errors.New("rpc response missing result")
	}
	if err := json.Unmarshal(envelope.Result, target); err != nil {
		return fmt.Errorf("decode rpc result: %w", err)
	}
	return nil
}

func cleanRPCURLs(values []string) []string {
	urls := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			urls = append(urls, trimmed)
		}
	}
	return urls
}

func normalizeAddress(value string) (string, error) {
	address := strings.ToLower(strings.TrimSpace(value))
	if !strings.HasPrefix(address, "0x") {
		address = "0x" + address
	}
	if len(address) != 42 {
		return "", fmt.Errorf("address must be 20-byte hex")
	}
	for _, r := range address[2:] {
		if !isHex(r) {
			return "", fmt.Errorf("address must be hex")
		}
	}
	return address, nil
}

func addressFromTopic(topic string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(topic))
	if !strings.HasPrefix(value, "0x") || len(value) != 66 {
		return "", fmt.Errorf("topic must be 32-byte hex")
	}
	for _, r := range value[2:] {
		if !isHex(r) {
			return "", fmt.Errorf("topic must be hex")
		}
	}
	return "0x" + value[len(value)-40:], nil
}

func uint256FromHex(value string) (*big.Int, error) {
	hexValue := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(value)), "0x")
	if hexValue == "" {
		hexValue = "0"
	}
	if len(hexValue) > 64 {
		return nil, fmt.Errorf("uint256 exceeds 32 bytes")
	}
	for _, r := range hexValue {
		if !isHex(r) {
			return nil, fmt.Errorf("uint256 must be hex")
		}
	}
	amount, ok := new(big.Int).SetString(hexValue, 16)
	if !ok {
		return nil, fmt.Errorf("parse uint256")
	}
	return amount, nil
}

func parseHexUint64(value string) (uint64, error) {
	hexValue := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(value)), "0x")
	if hexValue == "" {
		return 0, fmt.Errorf("empty hex value")
	}
	for _, r := range hexValue {
		if !isHex(r) {
			return 0, fmt.Errorf("value must be hex")
		}
	}
	return strconv.ParseUint(hexValue, 16, 64)
}

func isHex(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}
