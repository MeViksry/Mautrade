package workers

import (
	"math/big"
	"testing"

	"github.com/MeViksry/qdecimal"
)

func TestTokenAmountDecimalUsesPositiveTokenScale(t *testing.T) {
	t.Parallel()

	units := new(big.Int).Mul(big.NewInt(520), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	amount, err := tokenAmountDecimal(units, 18)
	if err != nil {
		t.Fatalf("expected token amount conversion to succeed, got %v", err)
	}
	if !amount.Equal(qdecimal.MustParse("520")) {
		t.Fatalf("expected 520 USDT, got %s", amount.String())
	}
}
