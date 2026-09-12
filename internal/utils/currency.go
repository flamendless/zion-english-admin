package utils

import (
	"fmt"
	"strconv"
	"strings"
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
