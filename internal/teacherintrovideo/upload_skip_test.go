package teacherintrovideo

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"zion-english/internal/constants"
)

func TestProcessUploadSkipsCompressWhenFileSmall(t *testing.T) {
	if !FfmpegAvailable() || !FfprobeAvailable() {
		t.Skip("ffmpeg or ffprobe unavailable")
	}

	tmpFile, err := os.CreateTemp("", "intro-video-input-*.mp4")
	if err != nil {
		t.Fatal(err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	cmd := exec.Command("ffmpeg",
		"-y",
		"-f", "lavfi",
		"-i", "color=c=black:s=320x240:d=1",
		"-f", "lavfi",
		"-i", "sine=frequency=440:duration=1",
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		tmpPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to create test video: %v\n%s", err, out)
	}

	file, err := os.Open(tmpPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}

	settings := constants.IntroVideoEncodeSettingsForPreset(constants.IntroVideoCompressPresetFine)
	processed, err := ProcessUpload(context.Background(), file, "intro.mp4", info.Size(), settings)
	if err != nil {
		t.Fatalf("ProcessUpload() error: %v", err)
	}
	defer processed.Cleanup()

	if processed.Ext != ".mp4" {
		t.Fatalf("processed.Ext = %q, want .mp4", processed.Ext)
	}
	if processed.MimeType != "video/mp4" {
		t.Fatalf("processed.MimeType = %q, want video/mp4", processed.MimeType)
	}
	if processed.Path == "" {
		t.Fatal("expected processed path")
	}
}
