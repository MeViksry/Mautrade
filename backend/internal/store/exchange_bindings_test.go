package store

import (
	"errors"
	"testing"
)

func TestNormalizeBindingStatusRevocationAliases(t *testing.T) {
	t.Parallel()

	tests := []string{"revoked", "disconnected", "disconnect", "deleted", "delete"}
	for _, status := range tests {
		status := status
		t.Run(status, func(t *testing.T) {
			t.Parallel()
			normalized, err := normalizeBindingStatus(status)
			if err != nil {
				t.Fatalf("expected valid status, got %v", err)
			}
			if normalized != "revoked" {
				t.Fatalf("expected revoked, got %s", normalized)
			}
		})
	}
}

func TestNormalizeBindingStatusRejectsUnknownValue(t *testing.T) {
	t.Parallel()

	_, err := normalizeBindingStatus("paused")
	if !errors.Is(err, ErrInvalidExchangeStatus) {
		t.Fatalf("expected ErrInvalidExchangeStatus, got %v", err)
	}
}
