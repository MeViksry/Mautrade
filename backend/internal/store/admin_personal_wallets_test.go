package store

import (
	"errors"
	"testing"
)

func TestNormalizePersonalWalletCode(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"viksry":       "viksry",
		"VIKSRY":       "viksry",
		"aryanto-hong": "aryanto_hong",
		"aryanto_hong": "aryanto_hong",
	}
	for input, expected := range tests {
		input := input
		expected := expected
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			if actual := normalizePersonalWalletCode(input); actual != expected {
				t.Fatalf("expected %s, got %s", expected, actual)
			}
		})
	}
}

func TestNormalizePersonalWalletCodeRejectsUnknown(t *testing.T) {
	t.Parallel()

	if actual := normalizePersonalWalletCode("finance"); actual != "" {
		t.Fatalf("expected empty invalid code, got %s", actual)
	}
}

func TestNormalizePersonalWalletAddress(t *testing.T) {
	t.Parallel()

	address, err := normalizePersonalWalletAddress(" 0x55d398326f99059fF775485246999027B3197955 ")
	if err != nil {
		t.Fatalf("expected valid address: %v", err)
	}
	if address != "0x55d398326f99059ff775485246999027b3197955" {
		t.Fatalf("unexpected normalized address %s", address)
	}
}

func TestNormalizePersonalWalletAddressAllowsEmpty(t *testing.T) {
	t.Parallel()

	address, err := normalizePersonalWalletAddress("  ")
	if err != nil {
		t.Fatalf("expected empty address to clear wallet: %v", err)
	}
	if address != "" {
		t.Fatalf("expected empty address, got %s", address)
	}
}

func TestNormalizePersonalWalletAddressRejectsMalformed(t *testing.T) {
	t.Parallel()

	_, err := normalizePersonalWalletAddress("0xnot-a-wallet")
	if !errors.Is(err, ErrInvalidPersonalWalletAddress) {
		t.Fatalf("expected ErrInvalidPersonalWalletAddress, got %v", err)
	}
}
