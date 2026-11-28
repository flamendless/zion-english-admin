package parentpaymentocr

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const defaultTesseractLanguages = "kor+eng"

// TesseractInstalled reports whether the tesseract binary is on PATH.
func TesseractInstalled() bool {
	_, err := exec.LookPath("tesseract")
	return err == nil
}

// ResolvedTesseractLanguages returns TESSERACT_LANGUAGES or the default kor+eng pack list.
func ResolvedTesseractLanguages() string {
	lang := strings.TrimSpace(os.Getenv("TESSERACT_LANGUAGES"))
	if lang == "" {
		return defaultTesseractLanguages
	}
	return lang
}

// DetectDocumentText runs Tesseract OCR on image bytes (PNG or JPEG).
func DetectDocumentText(ctx context.Context, image []byte, fileExt string) (string, error) {
	if _, err := exec.LookPath("tesseract"); err != nil {
		return "", ErrTesseractNotInstalled
	}
	if len(image) == 0 {
		return "", ErrOCRNoText
	}

	ext := strings.ToLower(strings.TrimSpace(fileExt))
	if ext == "" {
		ext = ".png"
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	switch ext {
	case ".png", ".jpg", ".jpeg":
	default:
		return "", fmt.Errorf("%w: unsupported image type %s", ErrOCRFailed, ext)
	}

	tmpDir, err := os.MkdirTemp("", "parent-payment-ocr-")
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrOCRFailed, err)
	}
	defer os.RemoveAll(tmpDir)

	imagePath := filepath.Join(tmpDir, "input"+ext)
	if err := os.WriteFile(imagePath, image, 0600); err != nil {
		return "", fmt.Errorf("%w: %v", ErrOCRFailed, err)
	}

	lang := ResolvedTesseractLanguages()

	// PSM 4: single column — fits stacked mobile banking transaction lists.
	cmd := exec.CommandContext(ctx, "tesseract", imagePath, "stdout", "-l", lang, "--psm", "4")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%w: %s", ErrOCRFailed, msg)
	}

	text := stripTesseractNoise(strings.TrimSpace(string(out)))
	if text == "" {
		return "", ErrOCRNoText
	}
	return text, nil
}

var tesseractDiagLineRE = regexp.MustCompile(`(?i)^Estimating resolution as \d+`)

// stripTesseractNoise removes rare diagnostic lines when stderr was merged into pasted text.
func stripTesseractNoise(text string) string {
	lines := strings.Split(text, "\n")
	out := lines[:0]
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || tesseractDiagLineRE.MatchString(trimmed) {
			continue
		}
		out = append(out, line)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}
