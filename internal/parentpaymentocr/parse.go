package parentpaymentocr

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var (
	lineAmountRE = regexp.MustCompile(`(?i)(?:^|[\s:：])([₩￥]?\s*[\d,]+(?:\.\d{1,2})?)\s*(?:KRW|krw|원|won)?\s*$`)
	amountOnlyRE = regexp.MustCompile(`^[\d,.\s₩￥원]+$`)
)

// ParsedLine is one legal name and payment amount extracted from OCR text.
type ParsedLine struct {
	LegalName string
	Amount    float64
	RawLine   string
}

// ParsePaymentLines scans OCR output for lines that pair a name with a payment amount.
func ParsePaymentLines(text string) []ParsedLine {
	var out []ParsedLine
	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || isLikelyHeader(line) {
			continue
		}
		if amountOnlyRE.MatchString(line) {
			continue
		}
		loc := lineAmountRE.FindStringSubmatchIndex(line)
		if loc == nil {
			continue
		}
		amountStr := strings.TrimSpace(line[loc[2]:loc[3]])
		name := strings.TrimSpace(line[:loc[0]])
		name = strings.TrimRight(name, ":：-|")
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		amount, ok := parseAmount(amountStr)
		if !ok || amount <= 0 {
			continue
		}
		out = append(out, ParsedLine{
			LegalName: name,
			Amount:    amount,
			RawLine:   line,
		})
	}
	return out
}

func isLikelyHeader(line string) bool {
	lower := strings.ToLower(line)
	switch {
	case strings.Contains(lower, "legal name"),
		strings.Contains(lower, "student name"),
		strings.Contains(lower, "amount"),
		strings.Contains(lower, "payment"),
		strings.Contains(lower, "total"):
		return countLetters(line) < 4
	default:
		return false
	}
}

func countLetters(s string) int {
	n := 0
	for _, r := range s {
		if unicode.IsLetter(r) {
			n++
		}
	}
	return n
}

func parseAmount(raw string) (float64, bool) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "₩")
	raw = strings.TrimPrefix(raw, "￥")
	raw = strings.ReplaceAll(raw, ",", "")
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

// AmountsMatch reports whether OCR and expected amounts are equal within one cent/won.
func AmountsMatch(parsed, expected float64) bool {
	if expected <= 0 && parsed <= 0 {
		return true
	}
	return math.Abs(parsed-expected) < 0.01
}

// NormalizeLegalName trims and collapses whitespace for lookup keys.
func NormalizeLegalName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	fields := strings.Fields(name)
	return strings.Join(fields, " ")
}
