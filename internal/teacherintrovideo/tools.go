package teacherintrovideo

import "os/exec"

func FfmpegAvailable() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

func FfprobeAvailable() bool {
	_, err := exec.LookPath("ffprobe")
	return err == nil
}
