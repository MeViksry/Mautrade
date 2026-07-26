package bscscan

import (
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
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type TxReceiptResponse struct {
	Result *struct {
		Status string `json:"status"` // "0x1" for success, "0x0" for failure
	} `json:"result"`
}

type TxByHashResponse struct {
	Result *struct {
		To    string `json:"to"`
		Input string `json:"input"`
	} `json:"result"`
}

// VerifyUSDTTransfer verifies if a given TXID is a successful USDT BEP-20 transfer to the specified wallet, and returns the amount transferred (in 18 decimals).
func (c *Client) VerifyUSDTTransfer(ctx context.Context, txID, expectedRecipient string) (*big.Int, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("bscscan: api key is empty")
	}
	expectedRecipient = strings.ToLower(strings.TrimSpace(expectedRecipient))
	expectedRecipient = strings.TrimPrefix(expectedRecipient, "0x")
	
	txID = strings.TrimSpace(txID)
	if !strings.HasPrefix(txID, "0x") {
		txID = "0x" + txID
	}

	// 1. Check receipt to ensure it was successful
	receiptURL := fmt.Sprintf("https://api.bscscan.com/api?module=proxy&action=eth_getTransactionReceipt&txhash=%s&apikey=%s", txID, c.apiKey)
	req1, _ := http.NewRequestWithContext(ctx, http.MethodGet, receiptURL, nil)
	res1, err := c.httpClient.Do(req1)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetwork, err)
	}
	defer res1.Body.Close()

	var receiptData TxReceiptResponse
	if err := json.NewDecoder(res1.Body).Decode(&receiptData); err != nil {
		return nil, fmt.Errorf("%w: decode receipt: %v", ErrNetwork, err)
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
	txURL := fmt.Sprintf("https://api.bscscan.com/api?module=proxy&action=eth_getTransactionByHash&txhash=%s&apikey=%s", txID, c.apiKey)
	req2, _ := http.NewRequestWithContext(ctx, http.MethodGet, txURL, nil)
	res2, err := c.httpClient.Do(req2)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetwork, err)
	}
	defer res2.Body.Close()

	var txData TxByHashResponse
	if err := json.NewDecoder(res2.Body).Decode(&txData); err != nil {
		return nil, fmt.Errorf("%w: decode tx: %v", ErrNetwork, err)
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

	// input is: method_id (10 chars) + recipient (64 chars) + amount (64 chars)
	// Example: 0xa9059cbb 000000000000000000000000 1234567890123456789012345678901234567890 00000...
	recipientData := input[10:74]
	// Address is the last 40 chars of the 64 char recipient block
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
