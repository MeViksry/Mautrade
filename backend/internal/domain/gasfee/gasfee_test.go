package gasfee

import (
	"testing"

	"github.com/MeViksry/qdecimal"
)

func TestCalculatorProfitSharePRDExample(t *testing.T) {
	t.Parallel()

	calculator := MustCalculator("0.5")
	result := calculator.CalculateFromValues(qdecimal.MustParse("100"), qdecimal.MustParse("150"))

	assertDecimal(t, result.GrossPnL, "50")
	assertDecimal(t, result.GasFeeAmount, "25.0")
	assertDecimal(t, result.NetAmountToUser, "25.0")
	if result.Type != TypeProfitShare {
		t.Fatalf("expected profit share, got %s", result.Type)
	}
}

func TestCalculatorLossRebatePRDExample(t *testing.T) {
	t.Parallel()

	calculator := MustCalculator("0.5")
	result := calculator.CalculateFromValues(qdecimal.MustParse("100"), qdecimal.MustParse("80"))

	assertDecimal(t, result.GrossPnL, "-20")
	assertDecimal(t, result.GasFeeAmount, "-10.0")
	assertDecimal(t, result.PlatformRebate, "10.0")
	assertDecimal(t, result.NetLossToUser, "10.0")
	if result.Type != TypeLossRebate {
		t.Fatalf("expected loss rebate, got %s", result.Type)
	}
}

func TestCalculatorHighPrecisionProfit(t *testing.T) {
	t.Parallel()

	calculator := MustCalculator("0.5")
	result := calculator.CalculateFromValues(
		qdecimal.MustParse("100.000000000000000000"),
		qdecimal.MustParse("150.123456789000000000"),
	)

	assertDecimal(t, result.GrossPnL, "50.123456789000000000")
	assertDecimal(t, result.GasFeeAmount, "25.0617283945000000000")
	assertDecimal(t, result.NetAmountToUser, "25.0617283945000000000")
}

func TestCalculatorFromPricesUsesSoldQuantityCostBasis(t *testing.T) {
	t.Parallel()

	calculator := MustCalculator("0.5")
	result := calculator.CalculateFromPrices(
		qdecimal.MustParse("62450.00"),
		qdecimal.MustParse("63100.50"),
		qdecimal.MustParse("0.01000000"),
	)

	assertDecimal(t, result.GrossPnL, "6.5050000000")
	assertDecimal(t, result.GasFeeAmount, "3.25250000000")
}

func assertDecimal(t *testing.T, actual qdecimal.Decimal, expected string) {
	t.Helper()
	if !actual.Equal(qdecimal.MustParse(expected)) {
		t.Fatalf("expected %s, got %s", expected, actual.String())
	}
}
