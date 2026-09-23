package teacherintrovideo

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"zion-english/internal/constants"
)

var allowedIntroVideoExtensions = map[string]string{
	".mp4":  "video/mp4",
	".webm": "video/webm",
	".mov":  "video/quicktime",
	".avi":  "video/x-msvideo",
	".mkv":  "video/x-matroska",
	".ogv":  "video/ogg",
	".m4v":  "video/x-m4v",
	".3gp":  "video/3gpp",
}

func ValidateUpload(file io.ReadSeeker, filename string, size int64) (ext, mimeType string, err error) {
	if size <= 0 {
		return "", "", ErrFileEmpty
	}
	if size > constants.MaxIntroVideoBytes {
		return "", "", ErrFileTooLarge
	}

	ext = strings.ToLower(filepath.Ext(filename))
	mimeType, ok := allowedIntroVideoExtensions[ext]
	if !ok {
		return "", "", ErrUnsupportedFormat
	}

	tmpFile, err := os.CreateTemp("", "intro-video-*"+ext)
	if err != nil {
		return "", "", ErrUploadPrepareFailed
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := io.Copy(tmpFile, file); err != nil {
		tmpFile.Close()
		return "", "", ErrReadFailed
	}
	if err := tmpFile.Close(); err != nil {
		return "", "", ErrReadFailed
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", "", ErrReadFailed
	}

	duration, err := probeDurationSeconds(tmpPath)
	if err != nil {
		return "", "", err
	}
	if duration <= 0 {
		return "", "", ErrEmptyDuration
	}
	if duration > float64(constants.MaxIntroVideoDurationSeconds) {
		return "", "", ErrTooLong
	}

	return ext, mimeType, nil
}

func probeDurationSeconds(path string) (float64, error) {
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return 0, ErrFfprobeUnavailable
	}

	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0, ErrInvalidContent
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return 0, ErrEmptyDuration
	}

	duration, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %s", ErrInvalidContent, raw)
	}
	return duration, nil
}
