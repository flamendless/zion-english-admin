package parentpaymentocr

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const defaultTesseractLanguages = "kor+eng"

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

	lang := strings.TrimSpace(os.Getenv("TESSERACT_LANGUAGES"))
	if lang == "" {
		lang = defaultTesseractLanguages
	}

	cmd := exec.CommandContext(ctx, "tesseract", imagePath, "stdout", "-l", lang, "--psm", "6")
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%w: %s", ErrOCRFailed, msg)
	}

	text := strings.TrimSpace(string(out))
	if text == "" {
		return "", ErrOCRNoText
	}
	return text, nil
}
