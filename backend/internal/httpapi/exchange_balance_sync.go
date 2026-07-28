package httpapi

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MeViksry/Mautrade/backend/internal/platform/exchangebalance"
	"github.com/MeViksry/Mautrade/backend/internal/store"
)

const exchangeBalanceSyncStaleAfter = 45 * time.Second
const exchangeBalanceSyncTimeout = 10 * time.Second

func (s *Server) syncDueUserExchangeBalances(ctx context.Context, userID string) {
	if s == nil || s.exchangeBalance == nil || !s.store.Ready() {
		return
	}
	asset := s.defaultBalanceAsset()
	staleBefore := time.Now().UTC().Add(-exchangeBalanceSyncStaleAfter)
	bindings, err := s.store.DueActiveExchangeBindingCredentials(ctx, userID, asset, staleBefore)
	if err != nil {
		s.logger.Warn("load due exchange balance bindings", "user_id", userID, "error", err)
		return
	}
	for _, binding := range bindings {
		syncCtx, cancel := context.WithTimeout(ctx, exchangeBalanceSyncTimeout)
		err := s.syncExchangeBindingBalance(syncCtx, userID, binding)
		cancel()
		if err != nil {
			s.logger.Warn(
				"sync exchange balance",
				"user_id", userID,
				"binding_id", binding.ID,
				"exchange", binding.ExchangeName,
				"account_mode", binding.AccountMode,
				"error", err,
			)
		}
	}
}

func (s *Server) syncExchangeBindingBalance(ctx context.Context, userID string, binding store.ExchangeBindingCredentialCiphertext) error {
	apiKey, err := s.credentialEncryptor.OpenString(binding.APIKeyCiphertext)
	if err != nil {
		return fmt.Errorf("decrypt exchange api key: %w", err)
	}
	apiSecret, err := s.credentialEncryptor.OpenString(binding.APISecretCiphertext)
	if err != nil {
		return fmt.Errorf("decrypt exchange api secret: %w", err)
	}
	passphrase := ""
	if len(binding.APIPassphraseCiphertext) > 0 {
		passphrase, err = s.credentialEncryptor.OpenString(binding.APIPassphraseCiphertext)
		if err != nil {
			return fmt.Errorf("decrypt exchange api passphrase: %w", err)
		}
	}
	return s.syncExchangeBindingBalanceWithPlaintext(ctx, userID, binding.ID, binding.ExchangeName, binding.AccountMode, apiKey, apiSecret, passphrase)
}

func (s *Server) syncExchangeBindingBalanceWithPlaintext(ctx context.Context, userID, bindingID, exchangeName, accountMode, apiKey, apiSecret, passphrase string) error {
	asset := s.defaultBalanceAsset()
	balance, err := s.exchangeBalance.FetchSpotBalance(ctx, exchangebalance.Credentials{
		Exchange:    exchangeName,
		APIKey:      strings.TrimSpace(apiKey),
		APISecret:   strings.TrimSpace(apiSecret),
		Passphrase:  strings.TrimSpace(passphrase),
		AccountMode: exchangebalance.NormalizeAccountMode(accountMode),
		Asset:       asset,
	})
	if err != nil {
		return err
	}
	return s.store.RecordExchangeBalanceSnapshot(ctx, store.ExchangeBalanceSnapshotParams{
		UserID:            userID,
		ExchangeBindingID: bindingID,
		Asset:             balance.Asset,
		FreeAmount:        balance.Free,
		LockedAmount:      balance.Locked,
		CapturedAt:        time.Now().UTC(),
	})
}

func (s *Server) defaultBalanceAsset() string {
	asset := strings.ToUpper(strings.TrimSpace(s.config.DefaultCurrency))
	if asset == "" {
		return "USDT"
	}
	return asset
}
