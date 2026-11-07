package parentpaymentocr

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var (
	lineAmountRE     = regexp.MustCompile(`(?i)(?:^|[\s:：])(\+?\s*[₩￥]?\s*[\d,]+(?:\.\d{1,2})?)\s*(?:KRW|krw|원|won|#)?\s*$`)
	amountOnlyRE     = regexp.MustCompile(`^[\d,.\s₩￥원+#]+$`)
	bankCreditLineRE = regexp.MustCompile(`(?i)^\+\s*([\d,]+(?:\.\d+)?)\s*(?:원|won|#)?\s*$`)
	txMetaLineRE     = regexp.MustCompile(`^\d{2}\.\d{2}\s|\d{4}\.\d{2}\.\d{2}|전자금융|스마트출금|거래내역|잔액`)
)

// ParsedLine is one legal name and payment amount extracted from OCR text.
type ParsedLine struct {
	LegalName string
	Amount    float64
	RawLine   string
}

// ParsePaymentLines scans OCR output for lines that pair a name with a payment amount.
// Supports a single "name amount" line and mobile-banking layouts where the payer name
// is on one line and "+59,360원" is on the next.
func ParsePaymentLines(text string) []ParsedLine {
	var out []ParsedLine
	var pendingName string

	appendPair := func(name, raw string, amount float64) {
		name = strings.TrimSpace(name)
		if name == "" || amount <= 0 {
			return
		}
		out = append(out, ParsedLine{
			LegalName: name,
			Amount:    amount,
			RawLine:   raw,
		})
	}

	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || isLikelyHeader(line) || isBankSkipLine(line) {
			pendingName = ""
			continue
		}

		if m := bankCreditLineRE.FindStringSubmatch(line); m != nil {
			amount, ok := parseAmount(m[1])
			if ok && pendingName != "" {
				appendPair(pendingName, pendingName+" | "+line, amount)
				pendingName = ""
				continue
			}
		}

		if loc := lineAmountRE.FindStringSubmatchIndex(line); loc != nil {
			amountStr := strings.TrimSpace(line[loc[2]:loc[3]])
			name := strings.TrimSpace(line[:loc[0]])
			name = strings.TrimRight(name, ":：-|")
			name = strings.TrimSpace(name)
			amount, ok := parseAmount(amountStr)
			if ok && name != "" && !isBankSkipLine(name) {
				appendPair(name, line, amount)
				pendingName = ""
				continue
			}
		}

		if amountOnlyRE.MatchString(line) && !strings.HasPrefix(line, "+") {
			pendingName = ""
			continue
		}

		if looksLikePayerName(line) {
			pendingName = line
		}
	}
	return out
}

func isBankSkipLine(line string) bool {
	lower := strings.ToLower(line)
	if txMetaLineRE.MatchString(line) {
		return true
	}
	if strings.Contains(lower, "vivobook") || strings.Contains(lower, "거래내역조회") {
		return true
	}
	if strings.Contains(line, "~") && strings.Contains(line, ".") {
		return true
	}
	if strings.HasPrefix(line, "<") || strings.HasPrefix(line, "2026.") {
		return true
	}
	return false
}

func looksLikePayerName(line string) bool {
	if isBankSkipLine(line) {
		return false
	}
	if len(line) > 64 {
		return false
	}
	letters := countLetters(line)
	if letters < 2 {
		return false
	}
	digits := 0
	for _, r := range line {
		if unicode.IsDigit(r) {
			digits++
		}
	}
	return digits == 0 || (letters >= digits && digits <= 2)
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
	raw = strings.TrimPrefix(raw, "+")
	raw = strings.TrimPrefix(raw, "₩")
	raw = strings.TrimPrefix(raw, "￥")
	raw = strings.ReplaceAll(raw, ",", "")
	raw = strings.TrimSpace(raw)
	raw = strings.TrimSuffix(raw, "원")
	raw = strings.TrimSuffix(raw, "#")
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
