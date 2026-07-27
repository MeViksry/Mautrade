package workers

import (
	"context"
	"errors"
	"log/slog"
	"math/big"
	"strings"
	"time"

	"github.com/MeViksry/Mautrade/backend/internal/platform/bscrpc"
	"github.com/MeViksry/Mautrade/backend/internal/store"
	"github.com/MeViksry/qdecimal"
)

type VerifierConfig struct {
	WalletAddress    string
	RPCURLs          []string
	ChainID          uint64
	TokenContract    string
	TokenDecimals    int
	MinConfirmations uint64
	Interval         time.Duration
}

type Verifier struct {
	store         *store.DashboardStore
	client        *bscrpc.Client
	logger        *slog.Logger
	wallet        string
	tokenContract string
	tokenDecimals int
	chainID       uint64
	interval      time.Duration
}

func NewVerifier(st *store.DashboardStore, config VerifierConfig, logger *slog.Logger) *Verifier {
	interval := config.Interval
	if interval <= 0 {
		interval = 30 * time.Second
	}
	tokenDecimals := config.TokenDecimals
	if tokenDecimals <= 0 {
		tokenDecimals = 18
	}
	tokenContract := strings.ToLower(strings.TrimSpace(config.TokenContract))
	if tokenContract == "" {
		tokenContract = bscrpc.DefaultUSDTContractAddress
	}
	chainID := config.ChainID
	if chainID == 0 {
		chainID = 56
	}
	return &Verifier{
		store: st,
		client: bscrpc.NewClient(bscrpc.Config{
			RPCURLs:          config.RPCURLs,
			ChainID:          chainID,
			TokenContract:    tokenContract,
			TokenDecimals:    tokenDecimals,
			MinConfirmations: config.MinConfirmations,
		}),
		logger:        logger,
		wallet:        strings.ToLower(strings.TrimSpace(config.WalletAddress)),
		tokenContract: tokenContract,
		tokenDecimals: tokenDecimals,
		chainID:       chainID,
		interval:      interval,
	}
}

func (v *Verifier) Start(ctx context.Context) {
	ticker := time.NewTicker(v.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			v.processPending(ctx)
		}
	}
}

func (v *Verifier) processPending(ctx context.Context) {
	deposits, err := v.store.PendingGasFeeDeposits(ctx, 50)
	if err != nil {
		v.logger.Error("gasfee verifier: failed to fetch pending deposits", "error", err)
		return
	}

	if !looksLikeEVMAddress(v.wallet) {
		if len(deposits) > 0 {
			v.logger.Warn("gasfee verifier: cannot verify deposits because GAS_FEE_DEPOSIT_ADDRESS is not a BSC wallet address")
		}
		return
	}

	settings, err := v.store.GlobalSettings(ctx)
	if err != nil {
		v.logger.Error("gasfee verifier: failed to fetch settings", "error", err)
		return
	}

	for _, dep := range deposits {
		time.Sleep(500 * time.Millisecond)

		if dep.TxID == nil || *dep.TxID == "" {
			continue
		}
		if !strings.EqualFold(dep.Asset, "USDT") {
			v.rejectDeposit(ctx, dep.ID, "Only USDT BEP-20 gas fee deposits are supported.", nil, nil, nil)
			continue
		}

		txID := *dep.TxID
		result, err := v.client.VerifyUSDTTransfer(ctx, txID, v.wallet)
		if err != nil {
			v.logger.Info("gasfee verifier: tx verification failed", "deposit_id", dep.ID, "tx_id", txID, "error", err)

			if errors.Is(err, bscrpc.ErrNetwork) || errors.Is(err, bscrpc.ErrPending) {
				v.logger.Info("gasfee verifier: retryable network error, skipping for now", "deposit_id", dep.ID)
				continue
			}
			v.rejectDeposit(ctx, dep.ID, "Invalid transaction: "+err.Error(), nil, nil, nil)
			continue
		}

		amountDec, decErr := tokenAmountDecimal(result.Amount, v.tokenDecimals)
		if decErr != nil {
			v.logger.Error("gasfee verifier: invalid decimal conversion", "deposit_id", dep.ID, "error", decErr)
			continue
		}

		expectedDec, parseErr := qdecimal.Parse(dep.Amount)
		if parseErr != nil {
			v.logger.Error("gasfee verifier: invalid deposit amount in db", "deposit_id", dep.ID, "amount", dep.Amount)
			continue
		}

		actualAmountStr := amountDec.String()
		blockNumber := result.BlockNumber
		confirmations := result.Confirmations
		if amountDec.Cmp(settings.MinDepositUsdt) < 0 {
			v.logger.Warn("gasfee verifier: amount below minimum", "deposit_id", dep.ID, "tx_id", txID, "minimum", settings.MinDepositUsdt, "actual", amountDec)
			v.rejectDeposit(ctx, dep.ID, "Amount below minimum. Minimum "+settings.MinDepositUsdt.String()+" USDT but received "+actualAmountStr, &actualAmountStr, &blockNumber, &confirmations)
			continue
		}

		v.logger.Info("gasfee verifier: tx verified successfully", "deposit_id", dep.ID, "tx_id", txID, "actual_amount", actualAmountStr)
		_, updateErr := v.store.SystemUpdateGasFeeDepositStatus(ctx, store.SystemUpdateGasFeeDepositStatusParams{
			DepositID:      dep.ID,
			Status:         "confirmed",
			ResolutionNote: "Auto-verified via BSC RPC. User requested " + expectedDec.String() + " USDT; received " + actualAmountStr + " USDT.",
			ActualAmount:   &actualAmountStr,
			SenderAddress:  &result.FromAddress,
			BlockNumber:    &blockNumber,
			Confirmations:  &confirmations,
			ChainID:        int64(result.ChainID),
			TokenContract:  result.TokenContract,
			Network:        "BEP-20",
			Now:            time.Now().UTC(),
		})
		if updateErr != nil {
			v.logger.Error("gasfee verifier: failed to confirm deposit", "deposit_id", dep.ID, "error", updateErr)
		}
	}
}

func (v *Verifier) rejectDeposit(ctx context.Context, depositID, note string, actualAmount *string, blockNumber *uint64, confirmations *uint64) {
	_, updateErr := v.store.SystemUpdateGasFeeDepositStatus(ctx, store.SystemUpdateGasFeeDepositStatusParams{
		DepositID:      depositID,
		Status:         "rejected",
		ResolutionNote: note,
		ActualAmount:   actualAmount,
		BlockNumber:    blockNumber,
		Confirmations:  confirmations,
		ChainID:        int64(v.chainID),
		TokenContract:  v.tokenContract,
		Network:        "BEP-20",
		Now:            time.Now().UTC(),
	})
	if updateErr != nil {
		v.logger.Error("gasfee verifier: failed to reject deposit", "deposit_id", depositID, "error", updateErr)
	}
}

func looksLikeEVMAddress(value string) bool {
	address := strings.ToLower(strings.TrimSpace(value))
	if !strings.HasPrefix(address, "0x") || len(address) != 42 {
		return false
	}
	for _, r := range address[2:] {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return false
		}
	}
	return true
}

func tokenAmountDecimal(amount *big.Int, decimals int) (qdecimal.Decimal, error) {
	if decimals < 0 {
		decimals = 0
	}
	return qdecimal.NewFromBigInt(amount, int32(decimals))
}
