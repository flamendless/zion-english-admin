package teacherintrovideo

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
	"zion-english/internal/constants"
)

func compressToMP4(ctx context.Context, inputPath, outputPath string, settings constants.IntroVideoEncodeSettings) error {
	if !FfmpegAvailable() {
		return ErrFfmpegUnavailable
	}

	scale := fmt.Sprintf("scale='min(%d,iw)':-2", settings.MaxWidth)
	args := []string{
		"-y",
		"-i", inputPath,
		"-map", "0:v:0",
		"-map", "0:a:0?",
		"-c:v", "libx264",
		"-preset", settings.Preset,
		"-crf", settings.CRF,
		"-pix_fmt", "yuv420p",
		"-vf", scale,
		"-c:a", "aac",
		"-b:a", settings.AudioBitrate,
		"-movflags", "+faststart",
		"-map_metadata", "-1",
		outputPath,
	}

	if _, ok := ctx.Deadline(); !ok {
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(settings.TimeoutSecs)*time.Second)
		defer cancel()
		ctx = timeoutCtx
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return fmt.Errorf("%w: ffmpeg timed out after %ds", ErrCompressFailed, settings.TimeoutSecs)
	}
	if ctx.Err() != nil {
		return fmt.Errorf("%w: ffmpeg stopped (%v)", ErrCompressFailed, ctx.Err())
	}
	return fmt.Errorf("%w: %s", ErrCompressFailed, trimExecOutput(out))
}
