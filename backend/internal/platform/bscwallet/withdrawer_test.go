package bscwallet

import "testing"

func TestDecimalToTokenUnits(t *testing.T) {
	tests := []struct {
		name     string
		amount   string
		decimals int
		expected string
	}{
		{name: "whole USDT with 18 decimals", amount: "1", decimals: 18, expected: "1000000000000000000"},
		{name: "fractional USDT with 18 decimals", amount: "0.5", decimals: 18, expected: "500000000000000000"},
		{name: "full precision", amount: "1.000000000000000001", decimals: 18, expected: "1000000000000000001"},
		{name: "plus prefix", amount: "+10.25", decimals: 18, expected: "10250000000000000000"},
		{name: "six decimals", amount: "12.345678", decimals: 6, expected: "12345678"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := DecimalToTokenUnits(test.amount, test.decimals)
			if err != nil {
				t.Fatalf("DecimalToTokenUnits returned error: %v", err)
			}
			if actual.String() != test.expected {
				t.Fatalf("expected %s, got %s", test.expected, actual.String())
			}
		})
	}
}

func TestDecimalToTokenUnitsRejectsInvalidAmounts(t *testing.T) {
	tests := []string{
		"",
		"0",
		"0.0",
		".1",
		"1.",
		"1.0000000000000000001",
		"-1",
		"1e18",
		"abc",
	}

	for _, amount := range tests {
		t.Run(amount, func(t *testing.T) {
			if actual, err := DecimalToTokenUnits(amount, 18); err == nil {
				t.Fatalf("expected error, got %s", actual.String())
			}
		})
	}
}

func TestParseAddress(t *testing.T) {
	address, err := ParseAddress("55d398326f99059ff775485246999027b3197955")
	if err != nil {
		t.Fatalf("ParseAddress returned error: %v", err)
	}
	if address.Hex() != "0x55d398326f99059fF775485246999027B3197955" {
		t.Fatalf("unexpected checksum address: %s", address.Hex())
	}
}
