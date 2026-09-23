package teacherintrovideo

import (
	"errors"
	"testing"
	"zion-english/internal/constants"
)

func TestParseYouTubeURL(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantURL string
		wantErr error
	}{
		{
			name:    "watch url",
			raw:     "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			wantURL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		},
		{
			name:    "short url",
			raw:     "https://youtu.be/dQw4w9WgXcQ",
			wantURL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		},
		{
			name:    "embed url",
			raw:     "https://www.youtube.com/embed/dQw4w9WgXcQ",
			wantURL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		},
		{
			name:    "shorts url",
			raw:     "https://www.youtube.com/shorts/dQw4w9WgXcQ",
			wantURL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		},
		{
			name:    "empty",
			raw:     "   ",
			wantErr: ErrURLRequired,
		},
		{
			name:    "invalid host",
			raw:     "https://example.com/watch?v=dQw4w9WgXcQ",
			wantErr: ErrInvalidURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseURL(constants.TeacherIntroVideoSourceYouTube, tt.raw)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.URL != tt.wantURL {
				t.Fatalf("expected url %q, got %q", tt.wantURL, got.URL)
			}
			if got.SourceType != constants.TeacherIntroVideoSourceYouTube {
				t.Fatalf("expected youtube source type, got %q", got.SourceType)
			}
		})
	}
}

func TestParseGoogleDriveURL(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantURL string
		wantErr error
	}{
		{
			name:    "file path url",
			raw:     "https://drive.google.com/file/d/abc123XYZ/view?usp=sharing",
			wantURL: "https://drive.google.com/file/d/abc123XYZ/view",
		},
		{
			name:    "open id url",
			raw:     "https://drive.google.com/open?id=abc123XYZ",
			wantURL: "https://drive.google.com/file/d/abc123XYZ/view",
		},
		{
			name:    "folder url rejected",
			raw:     "https://drive.google.com/drive/folders/abc123XYZ",
			wantErr: ErrInvalidURL,
		},
		{
			name:    "empty",
			raw:     "",
			wantErr: ErrURLRequired,
		},
		{
			name:    "unsupported host",
			raw:     "https://example.com/file/d/abc123XYZ/view",
			wantErr: ErrInvalidURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseURL(constants.TeacherIntroVideoSourceGoogleDrive, tt.raw)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.URL != tt.wantURL {
				t.Fatalf("expected url %q, got %q", tt.wantURL, got.URL)
			}
			if got.SourceType != constants.TeacherIntroVideoSourceGoogleDrive {
				t.Fatalf("expected google drive source type, got %q", got.SourceType)
			}
		})
	}
}

func TestParseURLRejectsUnsupportedSourceType(t *testing.T) {
	_, err := ParseURL(constants.TeacherIntroVideoSourceType("vimeo"), "https://example.com")
	if !errors.Is(err, ErrUnsupportedSourceType) {
		t.Fatalf("expected ErrUnsupportedSourceType, got %v", err)
	}
}
