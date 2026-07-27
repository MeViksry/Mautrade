package store

import (
	"errors"
	"testing"
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
