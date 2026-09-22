package trainingmaterials

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var youtubeIDPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)youtube\.com/watch\?.*v=([a-zA-Z0-9_-]{11})`),
	regexp.MustCompile(`(?i)youtu\.be/([a-zA-Z0-9_-]{11})`),
	regexp.MustCompile(`(?i)youtube\.com/embed/([a-zA-Z0-9_-]{11})`),
	regexp.MustCompile(`(?i)youtube\.com/shorts/([a-zA-Z0-9_-]{11})`),
}

type ParsedURL struct {
	SourceType   SourceType
	VideoID      string
	EmbedURL     string
	ThumbnailURL string
}

func ParseURL(sourceType SourceType, rawURL string) (ParsedURL, error) {
	switch sourceType {
	case SourceYouTube:
		return parseYouTubeURL(rawURL)
	default:
		return ParsedURL{}, ErrUnsupportedSourceType
	}
}

func InferSourceType(rawURL string) (SourceType, error) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return "", ErrURLRequired
	}
	if _, err := parseYouTubeURL(trimmed); err == nil {
		return SourceYouTube, nil
	}
	return "", ErrInvalidURL
}

func parseYouTubeURL(rawURL string) (ParsedURL, error) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return ParsedURL{}, ErrURLRequired
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return ParsedURL{}, ErrInvalidURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ParsedURL{}, ErrInvalidURL
	}

	videoID := extractYouTubeVideoID(trimmed)
	if videoID == "" {
		return ParsedURL{}, ErrInvalidURL
	}

	embedURL := fmt.Sprintf("https://www.youtube.com/embed/%s?enablejsapi=1&rel=0", videoID)
	thumbnailURL := fmt.Sprintf("https://img.youtube.com/vi/%s/hqdefault.jpg", videoID)

	return ParsedURL{
		SourceType:   SourceYouTube,
		VideoID:      videoID,
		EmbedURL:     embedURL,
		ThumbnailURL: thumbnailURL,
	}, nil
}

func extractYouTubeVideoID(rawURL string) string {
	for _, pattern := range youtubeIDPatterns {
		if matches := pattern.FindStringSubmatch(rawURL); len(matches) == 2 {
			return matches[1]
		}
	}
	return ""
}
