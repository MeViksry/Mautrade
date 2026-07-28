package store

import (
	"errors"
	"testing"

	"github.com/MeViksry/qdecimal"
)

func TestNormalizeGasFeeTxID(t *testing.T) {
	t.Parallel()

	hash := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	normalized, err := NormalizeGasFeeTxID(hash)
	if err != nil {
		t.Fatalf("expected valid hash, got %v", err)
	}
	expected := "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if normalized != expected {
		t.Fatalf("expected %s, got %s", expected, normalized)
	}
}

func TestNormalizeGasFeeTxIDRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		txID string
		want error
	}{
		{name: "empty", txID: "", want: ErrGasFeeDepositTxIDRequired},
		{name: "short", txID: "0xabc", want: ErrGasFeeDepositTxIDInvalid},
		{name: "non hex", txID: "0xgggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg", want: ErrGasFeeDepositTxIDInvalid},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := NormalizeGasFeeTxID(tt.txID)
			if !errors.Is(err, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, err)
			}
		})
	}
}

func TestCalculateAdminWalletCommissionAllocations(t *testing.T) {
	t.Parallel()

	allocations, err := calculateAdminWalletCommissionAllocations(qdecimal.MustParse("500"))
	if err != nil {
		t.Fatalf("expected allocations: %v", err)
	}
	assertCommissionAllocation(t, allocations, "viksry", "0.10", "50.00")
	assertCommissionAllocation(t, allocations, "aryanto_hong", "0.90", "450.00")
	assertCommissionTotal(t, allocations, "500")
}

func TestCalculateAdminWalletCommissionAllocationsPreservesTokenPrecision(t *testing.T) {
	t.Parallel()

	amount := qdecimal.MustParse("1.000000000000000001")
	allocations, err := calculateAdminWalletCommissionAllocations(amount)
	if err != nil {
		t.Fatalf("expected allocations: %v", err)
	}
	assertCommissionAllocation(t, allocations, "viksry", "0.10", "0.100000000000000000")
	assertCommissionAllocation(t, allocations, "aryanto_hong", "0.90", "0.900000000000000001")
	assertCommissionTotal(t, allocations, amount.String())
}

func assertCommissionAllocation(t *testing.T, allocations []adminWalletCommissionAllocation, walletCode, shareRate, amount string) {
	t.Helper()

	for _, allocation := range allocations {
		if allocation.WalletCode != walletCode {
			continue
		}
		if !allocation.ShareRate.Equal(qdecimal.MustParse(shareRate)) {
			t.Fatalf("expected %s share rate %s, got %s", walletCode, shareRate, allocation.ShareRate.String())
		}
		if !allocation.Amount.Equal(qdecimal.MustParse(amount)) {
			t.Fatalf("expected %s amount %s, got %s", walletCode, amount, allocation.Amount.String())
		}
		return
	}
	t.Fatalf("missing commission allocation for %s", walletCode)
}

func assertCommissionTotal(t *testing.T, allocations []adminWalletCommissionAllocation, expected string) {
	t.Helper()

	total := qdecimal.Zero
	for _, allocation := range allocations {
		total = total.Add(allocation.Amount)
	}
	if !total.Equal(qdecimal.MustParse(expected)) {
		t.Fatalf("expected commission total %s, got %s", expected, total.String())
	}
}
