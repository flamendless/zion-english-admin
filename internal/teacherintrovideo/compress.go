package teacherintrovideo

import (
	"context"
	"fmt"
	"os/exec"
	"time"
	"zion-english/internal/constants"
)

func compressToMP4(ctx context.Context, inputPath, outputPath string) error {
	if !FfmpegAvailable() {
		return ErrFfmpegUnavailable
	}

	scale := fmt.Sprintf("scale='min(%d,iw)':-2", constants.IntroVideoEncodeMaxWidth)
	args := []string{
		"-y",
		"-i", inputPath,
		"-map", "0:v:0",
		"-map", "0:a:0?",
		"-c:v", "libx264",
		"-preset", constants.IntroVideoEncodePreset,
		"-crf", constants.IntroVideoEncodeCRF,
		"-pix_fmt", "yuv420p",
		"-vf", scale,
		"-c:a", "aac",
		"-b:a", constants.IntroVideoEncodeAudioBitrate,
		"-movflags", "+faststart",
		"-map_metadata", "-1",
		outputPath,
	}

	if _, ok := ctx.Deadline(); !ok {
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(constants.IntroVideoCompressTimeoutSeconds)*time.Second)
		defer cancel()
		ctx = timeoutCtx
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		if ctx.Err() != nil {
			return ErrCompressFailed
		}
		_ = out
		return ErrCompressFailed
	}
	return nil
}
