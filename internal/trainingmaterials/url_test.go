package trainingmaterials

import "testing"

func TestParseYouTubeURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		videoID string
	}{
		{
			name:    "watch url",
			url:     "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			videoID: "dQw4w9WgXcQ",
		},
		{
			name:    "short url",
			url:     "https://youtu.be/dQw4w9WgXcQ",
			videoID: "dQw4w9WgXcQ",
		},
		{
			name:    "shorts url",
			url:     "https://www.youtube.com/shorts/dQw4w9WgXcQ",
			videoID: "dQw4w9WgXcQ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseURL(SourceYouTube, tt.url)
			if err != nil {
				t.Fatalf("ParseURL() error = %v", err)
			}
			if parsed.VideoID != tt.videoID {
				t.Fatalf("VideoID = %q, want %q", parsed.VideoID, tt.videoID)
			}
			if parsed.EmbedURL == "" || parsed.ThumbnailURL == "" {
				t.Fatalf("expected embed and thumbnail URLs")
			}
		})
	}
}

func TestInferSourceTypeRejectsNonYouTube(t *testing.T) {
	_, err := InferSourceType("https://example.com/video")
	if err != ErrInvalidURL {
		t.Fatalf("expected ErrInvalidURL, got %v", err)
	}
}
