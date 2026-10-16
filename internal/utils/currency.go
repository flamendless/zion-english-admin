package utils

import (
	"fmt"
	"strconv"
	"strings"

	"zion-english/internal/constants"
)

func FormatCurrencyAmount(amount float64, currency string) string {
	if currency == "PHP" {
		return formatPHPAmount(amount)
	}
	return strconv.FormatFloat(amount, 'f', -1, 64)
}

func FormatCurrencyAmountFixed(amount float64, currency string) string {
	if currency == "PHP" {
		return formatPHPAmount(amount)
	}
	return fmt.Sprintf("%.2f", amount)
}

func FormatCurrency(amount float64, currency string) string {
	return fmt.Sprintf("%s %s", FormatCurrencyAmountFixed(amount, currency), currency)
}

func FormatCurrencyCompact(amount float64, currency string) string {
	return constants.CurrencySymbol(currency) + formatCompactCurrencyAmount(amount, currency)
}

func FormatRatePerDuration(amount float64, currency string, durationMinutes int) string {
	labeled := FormatCurrencyCompact(amount, currency)
	if durationMinutes <= 0 {
		return labeled
	}
	return fmt.Sprintf("%s/%d min.", labeled, durationMinutes)
}

func formatCompactCurrencyAmount(amount float64, currency string) string {
	s := FormatCurrencyAmountFixed(amount, currency)
	if strings.HasSuffix(s, ".00") {
		return s[:len(s)-3]
	}
	return s
}

func formatPHPAmount(amount float64) string {
	negative := amount < 0
	if negative {
		amount = -amount
	}

	parts := strings.Split(fmt.Sprintf("%.2f", amount), ".")
	intPart := parts[0]
	decPart := parts[1]

	var b strings.Builder
	if negative {
		b.WriteByte('-')
	}
	n := len(intPart)
	for i, c := range intPart {
		if i > 0 && (n-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	b.WriteByte('.')
	b.WriteString(decPart)
	return b.String()
}
