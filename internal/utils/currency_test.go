package utils

import "testing"

func TestFormatCurrencyAmountFixed_PHP(t *testing.T) {
	tests := []struct {
		amount   float64
		expected string
	}{
		{13775, "13,775.00"},
		{13775.5, "13,775.50"},
		{100, "100.00"},
		{0, "0.00"},
		{1234567.89, "1,234,567.89"},
		{-500.25, "-500.25"},
	}

	for _, tc := range tests {
		got := FormatCurrencyAmountFixed(tc.amount, "PHP")
		if got != tc.expected {
			t.Fatalf("FormatCurrencyAmountFixed(%v, PHP) = %q, want %q", tc.amount, got, tc.expected)
		}
	}
}

func TestFormatCurrencyAmountFixed_OtherCurrencies(t *testing.T) {
	got := FormatCurrencyAmountFixed(13775, "KRW")
	if got != "13775.00" {
		t.Fatalf("FormatCurrencyAmountFixed(13775, KRW) = %q, want %q", got, "13775.00")
	}
}

func TestFormatCurrencyAmount_PHP(t *testing.T) {
	got := FormatCurrencyAmount(350, "PHP")
	if got != "350.00" {
		t.Fatalf("FormatCurrencyAmount(350, PHP) = %q, want %q", got, "350.00")
	}
}

func TestFormatCurrencyAmount_OtherCurrencies(t *testing.T) {
	got := FormatCurrencyAmount(100, "KRW")
	if got != "100" {
		t.Fatalf("FormatCurrencyAmount(100, KRW) = %q, want %q", got, "100")
	}
}

func TestFormatCurrency(t *testing.T) {
	got := FormatCurrency(13775, "PHP")
	if got != "13,775.00 PHP" {
		t.Fatalf("FormatCurrency(13775, PHP) = %q, want %q", got, "13,775.00 PHP")
	}
}
