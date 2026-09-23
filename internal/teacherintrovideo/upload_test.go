package teacherintrovideo

import (
	"testing"
	"zion-english/internal/constants"
)

func TestMaxIntroVideoSizeMB(t *testing.T) {
	if constants.MaxIntroVideoSizeMB() != 200 {
		t.Fatalf("MaxIntroVideoSizeMB() = %d, want 200", constants.MaxIntroVideoSizeMB())
	}
}

func TestIntroVideoFileTooLargeMessage(t *testing.T) {
	msg := constants.IntroVideoFileTooLargeMessage()
	want := "[INTRO VIDEO] File is too large. Maximum size is 200 MB."
	if msg != want {
		t.Fatalf("IntroVideoFileTooLargeMessage() = %q, want %q", msg, want)
	}
}

func TestMaxIntroVideoDurationLabel(t *testing.T) {
	if constants.MaxIntroVideoDurationLabel() != "2 minutes" {
		t.Fatalf("MaxIntroVideoDurationLabel() = %q, want %q", constants.MaxIntroVideoDurationLabel(), "2 minutes")
	}
}

func TestIntroVideoTooLongMessage(t *testing.T) {
	msg := constants.IntroVideoTooLongMessage()
	want := "[INTRO VIDEO] Video is too long. Maximum duration is 2 minutes."
	if msg != want {
		t.Fatalf("IntroVideoTooLongMessage() = %q, want %q", msg, want)
	}
}
