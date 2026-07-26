package bscscan

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"
)

var (
	ErrInvalidTransaction = errors.New("bscscan: invalid transaction")
	ErrNetwork            = errors.New("bscscan: network or api error")
)

const (
	USDTContractAddress = "0x55d398326f99059ff775485246999027b3197955"
	TransferMethodID    = "0xa9059cbb"
	RPCEndpoint         = "https://bsc-dataseed.binance.org/"
)

type Client struct {
	// apiKey is kept for backwards compatibility but no longer strictly required for RPC.
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type jsonrpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
	ID      int    `json:"id"`
}

type txReceiptResponse struct {
	Result *struct {
		Status string `json:"status"` // "0x1" for success, "0x0" for failure
	} `json:"result"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type txByHashResponse struct {
	Result *struct {
		To    string `json:"to"`
		Input string `json:"input"`
	} `json:"result"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *Client) VerifyUSDTTransfer(ctx context.Context, txID, expectedRecipient string) (*big.Int, error) {
	expectedRecipient = strings.ToLower(strings.TrimSpace(expectedRecipient))
	expectedRecipient = strings.TrimPrefix(expectedRecipient, "0x")

	txID = strings.TrimSpace(txID)
	if !strings.HasPrefix(txID, "0x") {
		txID = "0x" + txID
	}
	if len(txID) != 66 {
		return nil, fmt.Errorf("%w: invalid txhash length (got %d)", ErrInvalidTransaction, len(txID))
	}

	// 1. Check receipt to ensure it was successful
	receiptReqBody, _ := json.Marshal(jsonrpcRequest{
		JSONRPC: "2.0",
		Method:  "eth_getTransactionReceipt",
		Params:  []any{txID},
		ID:      1,
	})

	req1, _ := http.NewRequestWithContext(ctx, http.MethodPost, RPCEndpoint, bytes.NewReader(receiptReqBody))
	req1.Header.Set("Content-Type", "application/json")
	res1, err := c.httpClient.Do(req1)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetwork, err)
	}
	defer res1.Body.Close()

	var receiptData txReceiptResponse
	if err := json.NewDecoder(res1.Body).Decode(&receiptData); err != nil {
		return nil, fmt.Errorf("%w: decode receipt: %v", ErrNetwork, err)
	}
	if receiptData.Error != nil {
		return nil, fmt.Errorf("%w: rpc error: %s", ErrNetwork, receiptData.Error.Message)
	}
	if receiptData.Result == nil {
		return nil, fmt.Errorf("%w: transaction receipt not found (not mined yet or invalid txid)", ErrNetwork)
	}
	if receiptData.Result.Status == "0x0" {
		return nil, fmt.Errorf("%w: transaction reverted on-chain (status 0x0)", ErrInvalidTransaction)
	}
	if receiptData.Result.Status != "0x1" {
		return nil, fmt.Errorf("%w: transaction pending or unknown status: %s", ErrNetwork, receiptData.Result.Status)
	}

	// 2. Check transaction details to ensure it's USDT and the correct amount/recipient
	txReqBody, _ := json.Marshal(jsonrpcRequest{
		JSONRPC: "2.0",
		Method:  "eth_getTransactionByHash",
		Params:  []any{txID},
		ID:      2,
	})
	req2, _ := http.NewRequestWithContext(ctx, http.MethodPost, RPCEndpoint, bytes.NewReader(txReqBody))
	req2.Header.Set("Content-Type", "application/json")
	res2, err := c.httpClient.Do(req2)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetwork, err)
	}
	defer res2.Body.Close()

	var txData txByHashResponse
	if err := json.NewDecoder(res2.Body).Decode(&txData); err != nil {
		return nil, fmt.Errorf("%w: decode tx: %v", ErrNetwork, err)
	}
	if txData.Error != nil {
		return nil, fmt.Errorf("%w: rpc error: %s", ErrNetwork, txData.Error.Message)
	}
	if txData.Result == nil {
		return nil, fmt.Errorf("%w: transaction details not found", ErrNetwork)
	}

	if strings.ToLower(txData.Result.To) != USDTContractAddress {
		return nil, fmt.Errorf("%w: transaction 'to' address is not the USDT contract (got %s)", ErrInvalidTransaction, txData.Result.To)
	}

	input := strings.ToLower(txData.Result.Input)
	if !strings.HasPrefix(input, TransferMethodID) || len(input) != 138 {
		return nil, fmt.Errorf("%w: not a standard ERC20 transfer (invalid input data length or method id)", ErrInvalidTransaction)
	}

	recipientData := input[10:74]
	recipientHex := recipientData[24:]

	if recipientHex != expectedRecipient {
		return nil, fmt.Errorf("%w: recipient mismatch (got %s, expected %s)", ErrInvalidTransaction, recipientHex, expectedRecipient)
	}

	amountData := input[74:]
	amountHex := strings.TrimLeft(amountData, "0")
	if amountHex == "" {
		amountHex = "0"
	}

	amount, ok := new(big.Int).SetString(amountHex, 16)
	if !ok {
		return nil, fmt.Errorf("%w: failed to parse amount from hex %s", ErrInvalidTransaction, amountHex)
	}

	return amount, nil
}
