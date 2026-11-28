package parentpaymentocr

import (
	"strings"
	"testing"
)

func TestParsePaymentLines(t *testing.T) {
	text := `Legal name    Amount
김영희 150,000
박철수 ₩200000 KRW
Lee Minho 17500.50
header only
99,999
홍길동: 120000원`

	lines := ParsePaymentLines(text)
	if len(lines) != 4 {
		t.Fatalf("got %d lines, want 4", len(lines))
	}
	if lines[0].LegalName != "김영희" || lines[0].Amount != 150000 {
		t.Fatalf("line 0: %+v", lines[0])
	}
	if lines[1].LegalName != "박철수" || lines[1].Amount != 200000 {
		t.Fatalf("line 1: %+v", lines[1])
	}
	if lines[2].LegalName != "Lee Minho" || lines[2].Amount != 17500.5 {
		t.Fatalf("line 2: %+v", lines[2])
	}
	if lines[3].LegalName != "홍길동" || lines[3].Amount != 120000 {
		t.Fatalf("line 3: %+v", lines[3])
	}
}

func TestParsePaymentLinesBankStatement(t *testing.T) {
	// Layout from a typical Korean mobile banking transaction history (name, then +amount line).
	text := `11.07 07:08:07 | 전자금융
김현주
+59,360원
8,380,550
11.07 01:46:05 | 전자금융
원정윤
+85,500원
8.321.190원
11.06 17:26:28 | 전자금융
JENNY MOM
+30,000원
11.06 13:48:59 | 스마트출금
신정훈
+45,000원`

	lines := ParsePaymentLines(text)
	if len(lines) != 4 {
		t.Fatalf("got %d lines, want 4: %+v", len(lines), lines)
	}
	want := []struct {
		name   string
		amount float64
	}{
		{"김현주", 59360},
		{"원정윤", 85500},
		{"JENNY MOM", 30000},
		{"신정훈", 45000},
	}
	for i, w := range want {
		if lines[i].LegalName != w.name || lines[i].Amount != w.amount {
			t.Fatalf("line %d: got %+v want %s %.0f", i, lines[i], w.name, w.amount)
		}
	}
}

func TestParsePaymentEntriesIncomplete(t *testing.T) {
	text := `+59,360원
신정훈`

	entries := ParsePaymentEntries(text)
	var orphanAmount, nameOnly int
	for _, e := range entries {
		if e.HasAmount && !e.HasName {
			orphanAmount++
		}
		if e.HasName && !e.HasAmount {
			nameOnly++
		}
	}
	if orphanAmount != 1 || nameOnly != 1 {
		t.Fatalf("got %v", entries)
	}
}

func TestParsePaymentLinesBankStatementNoisyOCR(t *testing.T) {
	text := `Estimating resolution as 259
나 8:08 2 8 +

< 거래내역조회

안신혜
+45,000원
4,049,868 2
11.18 00:26:47 스마트입금
윤소정윤승우
+54,000원
4,004,868 2
11.18 00:19:13 전자금융
노소영
+36,000원
3,950,868원
11.17 22:13:17 전자금융
우예슬
+40,500원
3,914,.868원
11.17 21:34:24 전자금융
Lee Nayul
+27,000 A`

	entries := ParsePaymentEntries(text)
	if len(entries) < 5 {
		t.Fatalf("got %d entries, want at least 5: %+v", len(entries), entries)
	}
	lines := ParsePaymentLines(text)
	if len(lines) != 5 {
		t.Fatalf("got %d lines, want 5: %+v", len(lines), lines)
	}
	if !strings.Contains(entries[0].OtherDetails, "11.18 00:26:47") || !strings.Contains(entries[0].OtherDetails, "스마트입금") {
		t.Fatalf("entry 0 other details: %q", entries[0].OtherDetails)
	}
	if !strings.Contains(entries[0].OtherDetails, "4,049,868") {
		t.Fatalf("entry 0 balance missing: %q", entries[0].OtherDetails)
	}
	want := []struct {
		name   string
		amount float64
	}{
		{"안신혜", 45000},
		{"윤소정윤승우", 54000},
		{"노소영", 36000},
		{"우예슬", 40500},
		{"Lee Nayul", 27000},
	}
	for i, w := range want {
		if lines[i].LegalName != w.name || lines[i].Amount != w.amount {
			t.Fatalf("line %d: got %+v want %s %.0f", i, lines[i], w.name, w.amount)
		}
	}
}

func TestStripTesseractNoise(t *testing.T) {
	in := "Estimating resolution as 259\n안신혜\n+45,000원"
	got := stripTesseractNoise(in)
	if strings.Contains(got, "Estimating") {
		t.Fatalf("noise not stripped: %q", got)
	}
	if !strings.Contains(got, "안신혜") {
		t.Fatalf("content dropped: %q", got)
	}
}

func TestNormalizeLegalName(t *testing.T) {
	if got := NormalizeLegalName("  김  영희  "); got != "김 영희" {
		t.Fatalf("normalize: %q", got)
	}
}
