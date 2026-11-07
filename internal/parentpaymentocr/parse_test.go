package parentpaymentocr

import "testing"

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

func TestNormalizeLegalName(t *testing.T) {
	if got := NormalizeLegalName("  김  영희  "); got != "김 영희" {
		t.Fatalf("normalize: %q", got)
	}
}
