package workers

import (
	"context"
	"log/slog"
	"time"

	"github.com/MeViksry/Mautrade/backend/internal/platform/bscscan"
	"github.com/MeViksry/Mautrade/backend/internal/store"
	"github.com/MeViksry/qdecimal"
)

type Verifier struct {
	store  *store.DashboardStore
	client *bscscan.Client
	logger *slog.Logger
	wallet string
}

func NewVerifier(st *store.DashboardStore, apiKey, walletAddress string, logger *slog.Logger) *Verifier {
	return &Verifier{
		store:  st,
		client: bscscan.NewClient(apiKey),
		logger: logger,
		wallet: walletAddress,
	}
}

func (v *Verifier) Start(ctx context.Context) {
	if v.wallet == "" {
		v.logger.Warn("gasfee verifier: no central wallet address configured, skipping verification")
		return
	}
	
	ticker := time.NewTicker(30 * time.Second)
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
	// Fetch up to 50 pending deposits to verify
	deposits, err := v.store.PendingGasFeeDeposits(ctx, 50)
	if err != nil {
		v.logger.Error("gasfee verifier: failed to fetch pending deposits", "error", err)
		return
	}

	for _, dep := range deposits {
		// Rate limit: Max 5 calls/sec for Free Etherscan API.
		// VerifyUSDTTransfer makes 2 calls, so sleeping for 500ms allows 4 calls/sec.
		time.Sleep(500 * time.Millisecond)

		if dep.TxID == nil || *dep.TxID == "" {
			continue
		}

		txID := *dep.TxID
		amount, err := v.client.VerifyUSDTTransfer(ctx, txID, v.wallet)
		if err != nil {
			v.logger.Info("gasfee verifier: tx verification failed", "deposit_id", dep.ID, "tx_id", txID, "error", err)
			
			// Optional: We could mark it as failed immediately, or retry later.
			// Let's mark it as failed if it's explicitly rejected by BscScan.
			_, updateErr := v.store.SystemUpdateGasFeeDepositStatus(ctx, store.SystemUpdateGasFeeDepositStatusParams{
				DepositID:      dep.ID,
				Status:         "failed",
				ResolutionNote: "Invalid transaction or network error: " + err.Error(),
				Now:            time.Now().UTC(),
			})
			if updateErr != nil {
				v.logger.Error("gasfee verifier: failed to mark deposit as failed", "deposit_id", dep.ID, "error", updateErr)
			}
			continue
		}

		// Convert actual received amount to string (18 decimals for USDT BEP20)
		// Usually USDT on BEP20 has 18 decimals, so amount is in wei.
		// Wait! USDT on BSC actually has 18 decimals.
		// 500 USDT = 500 * 10^18.
		amountDec, decErr := qdecimal.NewFromBigInt(amount, -18)
		if decErr != nil {
			v.logger.Error("gasfee verifier: invalid decimal conversion", "deposit_id", dep.ID, "error", decErr)
			continue
		}
		
		expectedDec, err := qdecimal.Parse(dep.Amount)
		if err != nil {
			v.logger.Error("gasfee verifier: invalid deposit amount in db", "deposit_id", dep.ID, "amount", dep.Amount)
			continue
		}

		if amountDec.Cmp(expectedDec) < 0 {
			v.logger.Warn("gasfee verifier: amount too low", "deposit_id", dep.ID, "tx_id", txID, "expected", expectedDec, "actual", amountDec)
			_, updateErr := v.store.SystemUpdateGasFeeDepositStatus(ctx, store.SystemUpdateGasFeeDepositStatusParams{
				DepositID:      dep.ID,
				Status:         "failed",
				ResolutionNote: "Amount too low. Expected " + expectedDec.String() + " but received " + amountDec.String(),
				Now:            time.Now().UTC(),
			})
			if updateErr != nil {
				v.logger.Error("gasfee verifier: failed to update deposit", "deposit_id", dep.ID, "error", updateErr)
			}
			continue
		}

		// Success! Mark as confirmed
		v.logger.Info("gasfee verifier: tx verified successfully", "deposit_id", dep.ID, "tx_id", txID)
		_, updateErr := v.store.SystemUpdateGasFeeDepositStatus(ctx, store.SystemUpdateGasFeeDepositStatusParams{
			DepositID:      dep.ID,
			Status:         "confirmed",
			ResolutionNote: "Auto-verified via BscScan.",
			Now:            time.Now().UTC(),
		})
		if updateErr != nil {
			v.logger.Error("gasfee verifier: failed to confirm deposit", "deposit_id", dep.ID, "error", updateErr)
		}
	}
}
