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
		err := s.syncExchangeBindingBalanceAsset(syncCtx, userID, binding, asset)
		cancel()
		if err != nil {
			s.logger.Warn(
				"sync exchange balance",
				"user_id", userID,
				"binding_id", binding.ID,
				"exchange", binding.ExchangeName,
				"account_mode", binding.AccountMode,
				"asset", asset,
				"error", err,
			)
		}
	}
}

func (s *Server) syncDueAdminSignalExchangeBalances(ctx context.Context, params store.CreateSignalParams) {
	if s == nil || s.exchangeBalance == nil || !s.store.Ready() {
		return
	}
	asset := s.defaultBalanceAsset()
	if params.Type == "sell" {
		asset = baseAssetFromSignalSymbol(params.Symbol)
	}
	staleBefore := time.Now().UTC().Add(-exchangeBalanceSyncStaleAfter)
	targets, err := s.store.DueAllActiveExchangeBindingCredentials(ctx, asset, staleBefore)
	if err != nil {
		s.logger.Warn("load due exchange balance bindings for admin signal", "asset", asset, "error", err)
		return
	}
	for _, target := range targets {
		syncCtx, cancel := context.WithTimeout(ctx, exchangeBalanceSyncTimeout)
		err := s.syncExchangeBindingBalanceAsset(syncCtx, target.UserID, target.Binding, asset)
		cancel()
		if err != nil {
			s.logger.Warn(
				"sync exchange balance for admin signal",
				"user_id", target.UserID,
				"binding_id", target.Binding.ID,
				"exchange", target.Binding.ExchangeName,
				"account_mode", target.Binding.AccountMode,
				"asset", asset,
				"error", err,
			)
		}
	}
}

func (s *Server) syncExchangeBindingBalance(ctx context.Context, userID string, binding store.ExchangeBindingCredentialCiphertext) error {
	return s.syncExchangeBindingBalanceAsset(ctx, userID, binding, s.defaultBalanceAsset())
}

func (s *Server) syncExchangeBindingBalanceAsset(ctx context.Context, userID string, binding store.ExchangeBindingCredentialCiphertext, asset string) error {
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
	_, err = s.syncExchangeBindingBalanceAssetWithPlaintext(ctx, userID, binding.ID, binding.ExchangeName, binding.AccountMode, apiKey, apiSecret, passphrase, asset)
	return err
}

func (s *Server) syncExchangeBindingBalanceWithPlaintext(ctx context.Context, userID, bindingID, exchangeName, accountMode, apiKey, apiSecret, passphrase string) (string, error) {
	return s.syncExchangeBindingBalanceAssetWithPlaintext(ctx, userID, bindingID, exchangeName, accountMode, apiKey, apiSecret, passphrase, s.defaultBalanceAsset())
}

func (s *Server) syncExchangeBindingBalanceAssetWithPlaintext(ctx context.Context, userID, bindingID, exchangeName, accountMode, apiKey, apiSecret, passphrase, asset string) (string, error) {
	asset = strings.ToUpper(strings.TrimSpace(asset))
	if asset == "" {
		asset = s.defaultBalanceAsset()
	}
	balance, err := s.exchangeBalance.FetchSpotBalance(ctx, exchangebalance.Credentials{
		Exchange:    exchangeName,
		APIKey:      strings.TrimSpace(apiKey),
		APISecret:   strings.TrimSpace(apiSecret),
		Passphrase:  strings.TrimSpace(passphrase),
		AccountMode: exchangebalance.NormalizeAccountMode(accountMode),
		Asset:       asset,
	})
	if err != nil {
		return "", err
	}
	resolvedAccountMode := exchangebalance.NormalizeAccountMode(firstNonEmpty(balance.AccountMode, accountMode))
	if resolvedAccountMode == "" {
		resolvedAccountMode = exchangebalance.AccountModeReal
	}
	if err := s.store.RecordExchangeBalanceSnapshot(ctx, store.ExchangeBalanceSnapshotParams{
		UserID:            userID,
		ExchangeBindingID: bindingID,
		AccountMode:       resolvedAccountMode,
		Asset:             balance.Asset,
		FreeAmount:        balance.Free,
		LockedAmount:      balance.Locked,
		CapturedAt:        time.Now().UTC(),
	}); err != nil {
		return "", err
	}
	return resolvedAccountMode, nil
}

func baseAssetFromSignalSymbol(symbol string) string {
	normalized := strings.NewReplacer("-", "/", "_", "/").Replace(strings.ToUpper(strings.TrimSpace(symbol)))
	parts := strings.Split(normalized, "/")
	if len(parts) > 0 && strings.TrimSpace(parts[0]) != "" {
		return strings.TrimSpace(parts[0])
	}
	return "USDT"
}

func (s *Server) defaultBalanceAsset() string {
	asset := strings.ToUpper(strings.TrimSpace(s.config.DefaultCurrency))
	if asset == "" {
		return "USDT"
	}
	return asset
}
