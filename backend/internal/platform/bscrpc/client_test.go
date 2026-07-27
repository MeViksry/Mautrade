package bscrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"testing"
)

const (
	testTxHash    = "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	testRecipient = "0x1111111111111111111111111111111111111111"
	testSender    = "0x2222222222222222222222222222222222222222"
)

func TestVerifyUSDTTransferValidTransferSumsMatchingLogs(t *testing.T) {
	t.Parallel()

	client := newRPCClient(t, 5, func(method string) any {
		switch method {
		case "eth_chainId":
			return "0x38"
		case "eth_blockNumber":
			return "0x68"
		case "eth_getTransactionReceipt":
			return receipt("0x64", "0x1", []receiptLog{
				transferLog(DefaultUSDTContractAddress, testSender, testRecipient, usdtUnits("500")),
				transferLog(DefaultUSDTContractAddress, testSender, testRecipient, usdtUnits("20")),
			})
		default:
			t.Fatalf("unexpected method %s", method)
			return nil
		}
	})

	result, err := client.VerifyUSDTTransfer(context.Background(), strings.TrimPrefix(testTxHash, "0x"), testRecipient)
	if err != nil {
		t.Fatalf("expected valid transfer, got %v", err)
	}
	if result.TxID != testTxHash {
		t.Fatalf("expected normalized tx hash %s, got %s", testTxHash, result.TxID)
	}
	if result.Amount.String() != usdtUnits("520").String() {
		t.Fatalf("expected summed amount 520 USDT, got %s", result.Amount)
	}
	if result.BlockNumber != 100 {
		t.Fatalf("expected block 100, got %d", result.BlockNumber)
	}
	if result.Confirmations != 5 {
		t.Fatalf("expected 5 confirmations, got %d", result.Confirmations)
	}
	if result.FromAddress != testSender {
		t.Fatalf("expected sender %s, got %s", testSender, result.FromAddress)
	}
}

func TestVerifyUSDTTransferRejectsWrongToken(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, []receiptLog{
		transferLog("0x3333333333333333333333333333333333333333", testSender, testRecipient, usdtUnits("500")),
	}, "0x1", "0x64", "0x68", 5)

	_, err := client.VerifyUSDTTransfer(context.Background(), testTxHash, testRecipient)
	if !errors.Is(err, ErrInvalidTransaction) {
		t.Fatalf("expected invalid transaction, got %v", err)
	}
}

func TestVerifyUSDTTransferRejectsWrongRecipient(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, []receiptLog{
		transferLog(DefaultUSDTContractAddress, testSender, "0x3333333333333333333333333333333333333333", usdtUnits("500")),
	}, "0x1", "0x64", "0x68", 5)

	_, err := client.VerifyUSDTTransfer(context.Background(), testTxHash, testRecipient)
	if !errors.Is(err, ErrInvalidTransaction) {
		t.Fatalf("expected invalid transaction, got %v", err)
	}
}

func TestVerifyUSDTTransferRejectsRevertedTransaction(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, nil, "0x0", "0x64", "0x68", 5)

	_, err := client.VerifyUSDTTransfer(context.Background(), testTxHash, testRecipient)
	if !errors.Is(err, ErrInvalidTransaction) {
		t.Fatalf("expected invalid transaction, got %v", err)
	}
}

func TestVerifyUSDTTransferLeavesMissingReceiptPending(t *testing.T) {
	t.Parallel()

	client := newRPCClient(t, 1, func(method string) any {
		switch method {
		case "eth_chainId":
			return "0x38"
		case "eth_getTransactionReceipt":
			return nil
		default:
			t.Fatalf("unexpected method %s", method)
			return nil
		}
	})

	_, err := client.VerifyUSDTTransfer(context.Background(), testTxHash, testRecipient)
	if !errors.Is(err, ErrPending) {
		t.Fatalf("expected pending transaction, got %v", err)
	}
}

func TestVerifyUSDTTransferLeavesLowConfirmationsPending(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, []receiptLog{
		transferLog(DefaultUSDTContractAddress, testSender, testRecipient, usdtUnits("500")),
	}, "0x1", "0x64", "0x66", 5)

	_, err := client.VerifyUSDTTransfer(context.Background(), testTxHash, testRecipient)
	if !errors.Is(err, ErrPending) {
		t.Fatalf("expected pending transaction, got %v", err)
	}
}

func TestVerifyUSDTTransferRejectsMalformedTXID(t *testing.T) {
	t.Parallel()

	client := NewClient(Config{RPCURLs: []string{"http://127.0.0.1:1"}, ChainID: 56, TokenContract: DefaultUSDTContractAddress})
	_, err := client.VerifyUSDTTransfer(context.Background(), "not-a-hash", testRecipient)
	if !errors.Is(err, ErrInvalidTransaction) {
		t.Fatalf("expected invalid transaction, got %v", err)
	}
}

func newTestClient(t *testing.T, logs []receiptLog, status, receiptBlock, latestBlock string, minConfirmations uint64) *Client {
	t.Helper()
	return newRPCClient(t, minConfirmations, func(method string) any {
		switch method {
		case "eth_chainId":
			return "0x38"
		case "eth_blockNumber":
			return latestBlock
		case "eth_getTransactionReceipt":
			return receipt(receiptBlock, status, logs)
		default:
			t.Fatalf("unexpected method %s", method)
			return nil
		}
	})
}

func newRPCClient(t *testing.T, minConfirmations uint64, handler func(method string) any) *Client {
	t.Helper()
	return NewClient(Config{
		RPCURLs:          []string{"https://unit.test/rpc"},
		ChainID:          56,
		TokenContract:    DefaultUSDTContractAddress,
		MinConfirmations: minConfirmations,
		HTTPClient:       &http.Client{Transport: fakeRPCTransport{t: t, handler: handler}},
	})
}

type fakeRPCTransport struct {
	t       *testing.T
	handler func(method string) any
}

func (f fakeRPCTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var rpcReq jsonrpcRequest
	if err := json.NewDecoder(req.Body).Decode(&rpcReq); err != nil {
		f.t.Fatalf("decode request: %v", err)
	}
	body, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      rpcReq.ID,
		"result":  f.handler(rpcReq.Method),
	})
	if err != nil {
		f.t.Fatalf("encode response: %v", err)
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(string(body))),
		Request:    req,
	}, nil
}

func receipt(blockNumber, status string, logs []receiptLog) map[string]any {
	return map[string]any{
		"transactionHash": testTxHash,
		"status":          status,
		"blockNumber":     blockNumber,
		"from":            testSender,
		"to":              DefaultUSDTContractAddress,
		"logs":            logs,
	}
}

func transferLog(contract, from, to string, amount *big.Int) receiptLog {
	return receiptLog{
		Address: contract,
		Topics: []string{
			TransferEventTopic,
			topicAddress(from),
			topicAddress(to),
		},
		Data: uint256Hex(amount),
	}
}

func topicAddress(address string) string {
	return "0x" + strings.Repeat("0", 24) + strings.TrimPrefix(strings.ToLower(address), "0x")
}

func usdtUnits(amount string) *big.Int {
	base, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		panic("invalid amount")
	}
	return base.Mul(base, new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
}

func uint256Hex(value *big.Int) string {
	return fmt.Sprintf("0x%064x", value)
}
