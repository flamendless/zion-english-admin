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

func TestFormatCurrencyCompact(t *testing.T) {
	got := FormatCurrencyCompact(90, "PHP")
	if got != "₱90" {
		t.Fatalf("FormatCurrencyCompact(90, PHP) = %q, want %q", got, "₱90")
	}
}

func TestFormatRatePerDuration(t *testing.T) {
	tests := []struct {
		amount   float64
		currency string
		minutes  int
		expected string
	}{
		{90, "PHP", 25, "₱90/25 min."},
		{90.5, "PHP", 25, "₱90.50/25 min."},
		{13775, "KRW", 25, "₩13775/25 min."},
		{100, "PHP", 0, "₱100"},
	}
	for _, tc := range tests {
		got := FormatRatePerDuration(tc.amount, tc.currency, tc.minutes)
		if got != tc.expected {
			t.Fatalf("FormatRatePerDuration(%v, %q, %d) = %q, want %q", tc.amount, tc.currency, tc.minutes, got, tc.expected)
		}
	}
}
