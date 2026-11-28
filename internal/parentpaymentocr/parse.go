package parentpaymentocr

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var (
	lineAmountRE = regexp.MustCompile(`(?i)(?:^|[\s:：])(\+?\s*[₩￥]?\s*[\d,]+(?:\.\d{1,2})?)\s*(?:KRW|krw|원|won|#)?\s*$`)
	amountOnlyRE = regexp.MustCompile(`^[\d,.\s₩￥원+#]+$`)
	// Optional 1–2 char suffix catches misread 원 (e.g. "A", "21", trailing "2").
	bankCreditLineRE   = regexp.MustCompile(`(?i)^\+\s*([\d,]+(?:\.\d+)?)\s*(?:원|won|#)?(?:\s+(?:[A-Za-z#]{1,2}|\d{1,2}))?\s*$`)
	txMetaLineRE       = regexp.MustCompile(`^\d{2}\.\d{2}\s|\d{4}\.\d{2}\.\d{2}|전자금융|스마트출금|스마트입금|거래내역|잔액`)
	txDateTimePrefixRE = regexp.MustCompile(`^(\d{2}\.\d{2}\s+\d{2}:\d{2}:\d{2})\s*(?:[|·]\s*)?(.*)$`)
)

// ParsedLine is one legal name and payment amount extracted from OCR text.
type ParsedLine struct {
	LegalName string
	Amount    float64
	RawLine   string
}

// PaymentOCREntry is one row candidate from OCR, complete or partial.
type PaymentOCREntry struct {
	LegalName    string
	Amount       float64
	HasName      bool
	HasAmount    bool
	RawLine      string
	OtherDetails string
	BalanceAfter string
	TxDateTime   string
	TxChannel    string
	ScanFailed   bool
}

// Complete reports whether both name and a positive amount were read.
func (e PaymentOCREntry) Complete() bool {
	return e.HasName && e.HasAmount && e.Amount > 0 && strings.TrimSpace(e.LegalName) != ""
}

// ParsePaymentLines returns only complete name and amount pairs.
func ParsePaymentLines(text string) []ParsedLine {
	var out []ParsedLine
	for _, e := range ParsePaymentEntries(text) {
		if !e.Complete() {
			continue
		}
		out = append(out, ParsedLine{
			LegalName: e.LegalName,
			Amount:    e.Amount,
			RawLine:   e.RawLine,
		})
	}
	return out
}

// ParsePaymentEntries scans OCR output and returns every payment row candidate, including
// partial reads (name without amount, amount without name, or whole-scan failure markers).
func ParsePaymentEntries(text string) []PaymentOCREntry {
	var out []PaymentOCREntry
	var pendingName string

	appendEntry := func(e PaymentOCREntry) {
		out = append(out, e)
	}

	appendPairWithDetails := func(name, raw string, amount float64, lines []string, creditIdx int) int {
		name = strings.TrimSpace(name)
		if name == "" || amount <= 0 {
			return creditIdx
		}
		e := PaymentOCREntry{
			LegalName: name,
			Amount:    amount,
			HasName:   true,
			HasAmount: true,
			RawLine:   raw,
		}
		nextIdx := attachBankTransactionDetails(&e, lines, creditIdx)
		appendEntry(e)
		return nextIdx
	}

	lines := ocrTextLines(text)
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if isLikelyHeader(line) || isBankSkipLine(line) {
			pendingName = ""
			continue
		}

		if m := bankCreditLineRE.FindStringSubmatch(line); m != nil {
			amount, ok := parseAmount(m[1])
			if !ok || amount <= 0 {
				appendEntry(PaymentOCREntry{RawLine: line, HasAmount: false, HasName: false})
				pendingName = ""
				continue
			}
			if pendingName != "" {
				i = appendPairWithDetails(pendingName, pendingName+" | "+line, amount, lines, i)
				pendingName = ""
				continue
			}
			e := PaymentOCREntry{
				Amount:    amount,
				HasAmount: true,
				RawLine:   line,
			}
			i = attachBankTransactionDetails(&e, lines, i)
			appendEntry(e)
			continue
		}

		if amountOnlyRE.MatchString(line) && !strings.HasPrefix(line, "+") {
			pendingName = ""
			continue
		}

		if loc := lineAmountRE.FindStringSubmatchIndex(line); loc != nil {
			amountStr := strings.TrimSpace(line[loc[2]:loc[3]])
			name := strings.TrimSpace(line[:loc[0]])
			name = strings.TrimRight(name, ":：-|")
			name = strings.TrimSpace(name)
			amount, ok := parseAmount(amountStr)
			if ok && amount > 0 && name != "" && !isBankSkipLine(name) {
				i = appendPairWithDetails(name, line, amount, lines, i)
				pendingName = ""
				continue
			}
			if ok && amount > 0 {
				appendEntry(PaymentOCREntry{
					LegalName: name,
					Amount:    amount,
					HasName:   name != "",
					HasAmount: true,
					RawLine:   line,
				})
				pendingName = ""
				continue
			}
		}

		if looksLikePayerName(line) {
			pendingName = line
		}
	}

	if pendingName != "" {
		appendEntry(PaymentOCREntry{
			LegalName: pendingName,
			HasName:   true,
			RawLine:   pendingName,
		})
	}
	return out
}

func ocrTextLines(text string) []string {
	text = stripTesseractNoise(text)
	var lines []string
	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func attachBankTransactionDetails(e *PaymentOCREntry, lines []string, creditIdx int) int {
	i := creditIdx
	if i+1 < len(lines) && looksLikeBalanceLine(lines[i+1]) {
		e.BalanceAfter = strings.TrimSpace(lines[i+1])
		i++
	}
	if i+1 < len(lines) && looksLikeTxDetailLine(lines[i+1]) {
		e.TxDateTime, e.TxChannel = parseTxDetailLine(lines[i+1])
		i++
	}
	e.OtherDetails = formatOtherDetails(e.TxDateTime, e.TxChannel, e.BalanceAfter)
	return i
}

func looksLikeBalanceLine(line string) bool {
	if strings.HasPrefix(line, "+") {
		return false
	}
	if !amountOnlyRE.MatchString(line) {
		return false
	}
	digits := 0
	for _, r := range line {
		if unicode.IsDigit(r) {
			digits++
		}
	}
	return digits >= 6
}

func looksLikeTxDetailLine(line string) bool {
	return txDateTimePrefixRE.MatchString(line)
}

func parseTxDetailLine(line string) (dateTime, channel string) {
	m := txDateTimePrefixRE.FindStringSubmatch(line)
	if len(m) < 2 {
		return "", strings.TrimSpace(line)
	}
	dateTime = strings.TrimSpace(m[1])
	if len(m) >= 3 {
		channel = strings.TrimSpace(m[2])
	}
	return dateTime, channel
}

func formatOtherDetails(dateTime, channel, balance string) string {
	var parts []string
	if dateTime != "" {
		if channel != "" {
			parts = append(parts, dateTime+" "+channel)
		} else {
			parts = append(parts, dateTime)
		}
	} else if channel != "" {
		parts = append(parts, channel)
	}
	if balance != "" {
		parts = append(parts, balance)
	}
	return strings.Join(parts, "\n")
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
