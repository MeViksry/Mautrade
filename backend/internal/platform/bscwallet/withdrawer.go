package bscwallet

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	ErrSignerNotConfigured = errors.New("bscwallet: signer private key is not configured")
	ErrInvalidAddress      = errors.New("bscwallet: invalid EVM address")
	ErrInvalidAmount       = errors.New("bscwallet: invalid token amount")
	ErrBroadcastFailed     = errors.New("bscwallet: broadcast failed")
)

const (
	DefaultUSDTContractAddress = "0x55d398326f99059ff775485246999027b3197955"
	defaultGasLimit            = uint64(100_000)
	minTransferGasLimit        = uint64(65_000)
	maxTransferGasLimit        = uint64(300_000)
)

type Config struct {
	RPCURLs       []string
	ChainID       uint64
	TokenContract string
	TokenDecimals int
	PrivateKey    string
	GasLimit      uint64
}

type Withdrawer struct {
	rpcURLs       []string
	chainID       *big.Int
	tokenContract common.Address
	tokenDecimals int
	privateKey    *ecdsa.PrivateKey
	signerAddress common.Address
	gasLimit      uint64
}

type TransferResult struct {
	TxHash        string
	FromAddress   string
	ToAddress     string
	TokenContract string
	AmountAtomic  string
	GasPriceWei   string
	GasLimit      uint64
	Nonce         uint64
}

func NewWithdrawer(config Config) (*Withdrawer, error) {
	privateKey, err := parsePrivateKey(config.PrivateKey)
	if err != nil {
		return nil, err
	}

	tokenContractText := strings.TrimSpace(config.TokenContract)
	if tokenContractText == "" {
		tokenContractText = DefaultUSDTContractAddress
	}
	tokenContract, err := ParseAddress(tokenContractText)
	if err != nil {
		return nil, fmt.Errorf("%w: token contract: %v", ErrInvalidAddress, err)
	}

	rpcURLs := cleanRPCURLs(config.RPCURLs)
	if len(rpcURLs) == 0 {
		rpcURLs = []string{"https://bsc-dataseed.bnbchain.org"}
	}

	chainID := config.ChainID
	if chainID == 0 {
		chainID = 56
	}

	tokenDecimals := config.TokenDecimals
	if tokenDecimals == 0 {
		tokenDecimals = 18
	}
	if tokenDecimals < 0 || tokenDecimals > 36 {
		return nil, fmt.Errorf("%w: token decimals out of range", ErrInvalidAmount)
	}

	gasLimit := config.GasLimit
	if gasLimit == 0 {
		gasLimit = defaultGasLimit
	}

	return &Withdrawer{
		rpcURLs:       rpcURLs,
		chainID:       new(big.Int).SetUint64(chainID),
		tokenContract: tokenContract,
		tokenDecimals: tokenDecimals,
		privateKey:    privateKey,
		signerAddress: crypto.PubkeyToAddress(privateKey.PublicKey),
		gasLimit:      gasLimit,
	}, nil
}

func (w *Withdrawer) SignerAddress() string {
	if w == nil {
		return ""
	}
	return strings.ToLower(w.signerAddress.Hex())
}

func (w *Withdrawer) SendUSDTTransfer(ctx context.Context, recipient string, amount string) (TransferResult, error) {
	if w == nil {
		return TransferResult{}, ErrSignerNotConfigured
	}

	toAddress, err := ParseAddress(recipient)
	if err != nil {
		return TransferResult{}, fmt.Errorf("%w: recipient: %v", ErrInvalidAddress, err)
	}
	amountAtomic, err := DecimalToTokenUnits(amount, w.tokenDecimals)
	if err != nil {
		return TransferResult{}, err
	}
	data := erc20TransferData(toAddress, amountAtomic)

	var lastErr error
	for _, rpcURL := range w.rpcURLs {
		result, err := w.sendWithEndpoint(ctx, rpcURL, toAddress, amountAtomic, data)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = errors.New("no RPC endpoints configured")
	}
	return TransferResult{}, fmt.Errorf("%w: %v", ErrBroadcastFailed, lastErr)
}

func (w *Withdrawer) sendWithEndpoint(ctx context.Context, rpcURL string, toAddress common.Address, amountAtomic *big.Int, data []byte) (TransferResult, error) {
	callCtx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	client, err := ethclient.DialContext(callCtx, rpcURL)
	if err != nil {
		return TransferResult{}, fmt.Errorf("dial rpc: %w", err)
	}
	defer client.Close()

	chainID, err := client.ChainID(callCtx)
	if err != nil {
		return TransferResult{}, fmt.Errorf("read chain id: %w", err)
	}
	if chainID.Cmp(w.chainID) != 0 {
		return TransferResult{}, fmt.Errorf("chain id mismatch got %s expected %s", chainID.String(), w.chainID.String())
	}

	nonce, err := client.PendingNonceAt(callCtx, w.signerAddress)
	if err != nil {
		return TransferResult{}, fmt.Errorf("read nonce: %w", err)
	}
	gasPrice, err := client.SuggestGasPrice(callCtx)
	if err != nil {
		return TransferResult{}, fmt.Errorf("suggest gas price: %w", err)
	}
	gasLimit, err := w.estimateGasLimit(callCtx, client, data)
	if err != nil {
		return TransferResult{}, err
	}

	tx := types.NewTransaction(nonce, w.tokenContract, big.NewInt(0), gasLimit, gasPrice, data)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(w.chainID), w.privateKey)
	if err != nil {
		return TransferResult{}, fmt.Errorf("sign transaction: %w", err)
	}
	if err := client.SendTransaction(callCtx, signedTx); err != nil {
		return TransferResult{}, fmt.Errorf("send transaction: %w", err)
	}

	return TransferResult{
		TxHash:        strings.ToLower(signedTx.Hash().Hex()),
		FromAddress:   strings.ToLower(w.signerAddress.Hex()),
		ToAddress:     strings.ToLower(toAddress.Hex()),
		TokenContract: strings.ToLower(w.tokenContract.Hex()),
		AmountAtomic:  amountAtomic.String(),
		GasPriceWei:   gasPrice.String(),
		GasLimit:      gasLimit,
		Nonce:         nonce,
	}, nil
}

func (w *Withdrawer) estimateGasLimit(ctx context.Context, client *ethclient.Client, data []byte) (uint64, error) {
	estimated, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From:  w.signerAddress,
		To:    &w.tokenContract,
		Value: big.NewInt(0),
		Data:  data,
	})
	if err != nil {
		return 0, fmt.Errorf("estimate gas: %w", err)
	}

	buffered := estimated + estimated/5 + 5_000
	if buffered < minTransferGasLimit {
		buffered = minTransferGasLimit
	}
	if buffered > maxTransferGasLimit {
		buffered = maxTransferGasLimit
	}
	if w.gasLimit > 0 && buffered < w.gasLimit {
		buffered = w.gasLimit
	}
	return buffered, nil
}

func ParseAddress(value string) (common.Address, error) {
	address := strings.ToLower(strings.TrimSpace(value))
	if !strings.HasPrefix(address, "0x") {
		address = "0x" + address
	}
	if !common.IsHexAddress(address) {
		return common.Address{}, fmt.Errorf("address must be 20-byte hex")
	}
	return common.HexToAddress(address), nil
}

func DecimalToTokenUnits(value string, decimals int) (*big.Int, error) {
	if decimals < 0 || decimals > 36 {
		return nil, fmt.Errorf("%w: decimals out of range", ErrInvalidAmount)
	}

	amount := strings.TrimSpace(value)
	if strings.HasPrefix(amount, "+") {
		amount = strings.TrimPrefix(amount, "+")
	}
	if amount == "" {
		return nil, fmt.Errorf("%w: amount is required", ErrInvalidAmount)
	}

	parts := strings.Split(amount, ".")
	if len(parts) > 2 || parts[0] == "" {
		return nil, fmt.Errorf("%w: malformed amount", ErrInvalidAmount)
	}
	if !isDecimalDigits(parts[0]) {
		return nil, fmt.Errorf("%w: malformed integer amount", ErrInvalidAmount)
	}

	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
		if fraction == "" || !isDecimalDigits(fraction) {
			return nil, fmt.Errorf("%w: malformed fractional amount", ErrInvalidAmount)
		}
		if len(fraction) > decimals {
			return nil, fmt.Errorf("%w: amount exceeds token decimals", ErrInvalidAmount)
		}
	}

	atomicText := strings.TrimLeft(parts[0]+fraction+strings.Repeat("0", decimals-len(fraction)), "0")
	if atomicText == "" {
		atomicText = "0"
	}
	atomic, ok := new(big.Int).SetString(atomicText, 10)
	if !ok || atomic.Sign() <= 0 {
		return nil, fmt.Errorf("%w: amount must be greater than zero", ErrInvalidAmount)
	}
	return atomic, nil
}

func erc20TransferData(recipient common.Address, amount *big.Int) []byte {
	methodID := crypto.Keccak256([]byte("transfer(address,uint256)"))[:4]
	data := make([]byte, 0, 4+32+32)
	data = append(data, methodID...)
	data = append(data, common.LeftPadBytes(recipient.Bytes(), 32)...)
	data = append(data, common.LeftPadBytes(amount.Bytes(), 32)...)
	return data
}

func parsePrivateKey(value string) (*ecdsa.PrivateKey, error) {
	keyText := strings.TrimSpace(value)
	if keyText == "" {
		return nil, ErrSignerNotConfigured
	}
	keyText = strings.TrimPrefix(keyText, "0x")
	if len(keyText) != 64 || !isHex(keyText) {
		return nil, fmt.Errorf("bscwallet: GAS_FEE_WITHDRAW_PRIVATE_KEY must be 32-byte hex")
	}
	privateKey, err := crypto.HexToECDSA(keyText)
	if err != nil {
		return nil, fmt.Errorf("bscwallet: parse private key: %w", err)
	}
	return privateKey, nil
}

func withDefaultTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, 30*time.Second)
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

func isDecimalDigits(value string) bool {
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func isHex(value string) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') && (char < 'A' || char > 'F') {
			return false
		}
	}
	return true
}
