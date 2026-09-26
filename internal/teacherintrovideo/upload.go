package teacherintrovideo

import (
	"context"
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

type ProcessedUpload struct {
	Path         string
	OriginalSize int64
	Size         int64
	Ext          string
	MimeType     string
	Cleanup      func()
}

func ProcessUpload(ctx context.Context, file io.ReadSeeker, filename string, size int64, settings constants.IntroVideoEncodeSettings) (ProcessedUpload, error) {
	ext, mimeType, inputPath, cleanupInput, err := validateAndSaveUpload(file, filename, size)
	if err != nil {
		return ProcessedUpload{}, err
	}
	return processStagedPath(ctx, inputPath, ext, mimeType, cleanupInput, settings)
}

// StageUpload validates the upload, saves it to a temp file, and returns paths for async processing.
func StageUpload(file io.ReadSeeker, filename string, size int64) (ext, mimeType, tmpPath string, cleanup func(), err error) {
	return validateAndSaveUpload(file, filename, size)
}

// ProcessStagedFile runs compression on a file already saved to inputPath (from StageUpload).
func ProcessStagedFile(ctx context.Context, inputPath, ext, mimeType string, settings constants.IntroVideoEncodeSettings) (ProcessedUpload, error) {
	cleanupInput := func() { os.Remove(inputPath) }
	return processStagedPath(ctx, inputPath, ext, mimeType, cleanupInput, settings)
}

func processStagedPath(ctx context.Context, inputPath, ext, mimeType string, cleanupInput func(), settings constants.IntroVideoEncodeSettings) (ProcessedUpload, error) {
	inputInfo, err := os.Stat(inputPath)
	if err != nil {
		cleanupInput()
		return ProcessedUpload{}, ErrReadFailed
	}

	if !constants.IntroVideoShouldCompress(inputInfo.Size(), settings) {
		return ProcessedUpload{
			Path:         inputPath,
			OriginalSize: inputInfo.Size(),
			Size:         inputInfo.Size(),
			Ext:          ext,
			MimeType:     mimeType,
			Cleanup:      cleanupInput,
		}, nil
	}

	compressedPath, err := os.CreateTemp("", "intro-video-compressed-*"+constants.IntroVideoStoredExt)
	if err != nil {
		cleanupInput()
		return ProcessedUpload{}, ErrUploadPrepareFailed
	}
	compressedPath.Close()

	outputPath := compressedPath.Name()
	cleanupOutput := func() { os.Remove(outputPath) }

	timeoutSecs := constants.IntroVideoCompressTimeoutSecs(inputInfo.Size(), settings)
	if err := compressToMP4(ctx, inputPath, outputPath, settings, timeoutSecs); err != nil {
		cleanupInput()
		cleanupOutput()
		return ProcessedUpload{}, err
	}

	compressedInfo, err := os.Stat(outputPath)
	if err != nil {
		cleanupInput()
		cleanupOutput()
		return ProcessedUpload{}, ErrCompressFailed
	}

	finalPath := outputPath
	finalSize := compressedInfo.Size()
	finalExt := constants.IntroVideoStoredExt
	finalMIME := constants.IntroVideoStoredMIME
	cleanupFinal := func() {
		cleanupInput()
		cleanupOutput()
	}

	if compressedInfo.Size() >= inputInfo.Size() {
		finalPath = inputPath
		finalSize = inputInfo.Size()
		finalExt = ext
		finalMIME = mimeType
		cleanupFinal = func() {
			cleanupOutput()
			cleanupInput()
		}
	}

	return ProcessedUpload{
		Path:         finalPath,
		OriginalSize: inputInfo.Size(),
		Size:         finalSize,
		Ext:          finalExt,
		MimeType:     finalMIME,
		Cleanup:      cleanupFinal,
	}, nil
}

func validateAndSaveUpload(file io.ReadSeeker, filename string, size int64) (ext, mimeType, tmpPath string, cleanup func(), err error) {
	if size <= 0 {
		return "", "", "", nil, ErrFileEmpty
	}
	if size > constants.MaxIntroVideoBytes {
		return "", "", "", nil, ErrFileTooLarge
	}

	ext = strings.ToLower(filepath.Ext(filename))
	mimeType, ok := allowedIntroVideoExtensions[ext]
	if !ok {
		return "", "", "", nil, ErrUnsupportedFormat
	}

	tmpFile, err := os.CreateTemp("", "intro-video-*"+ext)
	if err != nil {
		return "", "", "", nil, ErrUploadPrepareFailed
	}
	tmpPath = tmpFile.Name()
	cleanup = func() { os.Remove(tmpPath) }

	if _, err := io.Copy(tmpFile, file); err != nil {
		tmpFile.Close()
		cleanup()
		return "", "", "", nil, ErrReadFailed
	}
	if err := tmpFile.Close(); err != nil {
		cleanup()
		return "", "", "", nil, ErrReadFailed
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		cleanup()
		return "", "", "", nil, ErrReadFailed
	}

	duration, err := probeDurationSeconds(tmpPath)
	if err != nil {
		cleanup()
		return "", "", "", nil, err
	}
	if duration <= 0 {
		cleanup()
		return "", "", "", nil, ErrEmptyDuration
	}
	if duration > float64(constants.MaxIntroVideoDurationSeconds) {
		cleanup()
		return "", "", "", nil, ErrTooLong
	}

	return ext, mimeType, tmpPath, cleanup, nil
}

func probeDurationSeconds(path string) (float64, error) {
	if !FfprobeAvailable() {
		return 0, ErrFfprobeUnavailable
	}

	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("%w: %s", ErrInvalidContent, trimExecOutput(out))
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
