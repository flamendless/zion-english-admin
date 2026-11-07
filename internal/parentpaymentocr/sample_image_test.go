package parentpaymentocr

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSampleBankingScreenshotOCR(t *testing.T) {
	if os.Getenv("TESSERACT_INTEGRATION") == "" {
		t.Skip("set TESSERACT_INTEGRATION=1 to run")
	}
	if _, err := exec.LookPath("tesseract"); err != nil {
		t.Skip("tesseract not installed")
	}

	paths := []string{
		filepath.Join("testdata", "parent_payment_bank_sample.jpg"),
	}
	var image []byte
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err == nil {
			image = b
			break
		}
	}
	if len(image) == 0 {
		t.Skip("sample banking image not found")
	}

	text, err := DetectDocumentText(context.Background(), image, ".jpg")
	if err != nil {
		t.Fatalf("ocr: %v", err)
	}
	lines := ParsePaymentLines(text)
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 payments from sample, got %d; ocr:\n%s", len(lines), text)
	}

	byName := map[string]float64{}
	for _, row := range lines {
		byName[row.LegalName] = row.Amount
	}
	if byName["김현주"] != 59360 {
		t.Fatalf("김현주: got %v", byName["김현주"])
	}
	if byName["원정윤"] != 85500 {
		t.Fatalf("원정윤: got %v", byName["원정윤"])
	}
	if byName["JENNY MOM"] != 30000 {
		t.Fatalf("JENNY MOM: got %v", byName["JENNY MOM"])
	}
}
