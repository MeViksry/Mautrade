package httpapi

import (
	"testing"
	"time"

	"github.com/MeViksry/Mautrade/backend/internal/store"
)

func TestExchangeBindingDeletedResponseDoesNotExposeCredentialState(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	response := exchangeBindingDeletedResponse(store.ExchangeBindingCredentialCiphertext{
		ID:                      "binding-id",
		ExchangeName:            "binance",
		AccountMode:             "real",
		Status:                  "revoked",
		APIKeyCiphertext:        []byte("old-api-key"),
		APISecretCiphertext:     []byte("old-api-secret"),
		APIPassphraseCiphertext: []byte("old-passphrase"),
		PermissionScope:         "trade_only",
		LastVerifiedAt:          &now,
		CreatedAt:               now,
		UpdatedAt:               now,
	})

	if response.Status != "revoked" {
		t.Fatalf("expected revoked status, got %s", response.Status)
	}
	if response.MaskedAPIKey != "" {
		t.Fatalf("expected empty masked api key, got %s", response.MaskedAPIKey)
	}
	if response.HasAPISecret || response.HasPassphrase || response.LastVerifiedAt != nil {
		t.Fatalf("expected deleted response to hide credential state: %+v", response)
	}
}

func TestExchangeBindingAccountModeCandidatesAutoDetectsBinanceTestnet(t *testing.T) {
	t.Parallel()

	candidates := exchangeBindingAccountModeCandidates("binance", "real", false)
	if len(candidates) != 2 || candidates[0] != "real" || candidates[1] != "testnet" {
		t.Fatalf("expected real then testnet candidates, got %#v", candidates)
	}
}

func TestExchangeBindingAccountModeCandidatesAutoDetectsBybitDemoAndTestnet(t *testing.T) {
	t.Parallel()

	candidates := exchangeBindingAccountModeCandidates("bybit", "real", false)
	if len(candidates) != 3 || candidates[0] != "real" || candidates[1] != "demo" || candidates[2] != "testnet" {
		t.Fatalf("expected real, demo, then testnet candidates, got %#v", candidates)
	}
}

func TestExchangeBindingAccountModeCandidatesTokocryptoSkipsDemo(t *testing.T) {
	t.Parallel()

	candidates := exchangeBindingAccountModeCandidates("tokocrypto", "real", false)
	if len(candidates) != 1 || candidates[0] != "real" {
		t.Fatalf("expected only real candidate, got %#v", candidates)
	}
}

func TestExchangeBindingAccountModeCandidatesHonorsExplicitMode(t *testing.T) {
	t.Parallel()

	candidates := exchangeBindingAccountModeCandidates("binance", "testnet", true)
	if len(candidates) != 1 || candidates[0] != "testnet" {
		t.Fatalf("expected explicit testnet only, got %#v", candidates)
	}
}

func TestNormalizeExchangeAccountModeValueAcceptsTestnet(t *testing.T) {
	t.Parallel()

	mode, err := normalizeExchangeAccountModeValue("sandbox")
	if err != nil {
		t.Fatalf("expected sandbox alias to be accepted: %v", err)
	}
	if mode != "testnet" {
		t.Fatalf("expected sandbox to normalize to testnet, got %s", mode)
	}
}
