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
