package teacherintrovideo

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"zion-english/internal/constants"
)

var youtubeIDPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)youtube\.com/watch\?.*v=([a-zA-Z0-9_-]{11})`),
	regexp.MustCompile(`(?i)youtu\.be/([a-zA-Z0-9_-]{11})`),
	regexp.MustCompile(`(?i)youtube\.com/embed/([a-zA-Z0-9_-]{11})`),
	regexp.MustCompile(`(?i)youtube\.com/shorts/([a-zA-Z0-9_-]{11})`),
}

var googleDriveFilePattern = regexp.MustCompile(`(?i)drive\.google\.com/file/d/([a-zA-Z0-9_-]+)`)

type ParsedURL struct {
	SourceType constants.TeacherIntroVideoSourceType
	URL        string
}

func ParseURL(sourceType constants.TeacherIntroVideoSourceType, rawURL string) (ParsedURL, error) {
	switch sourceType {
	case constants.TeacherIntroVideoSourceYouTube:
		return parseYouTubeURL(rawURL)
	case constants.TeacherIntroVideoSourceGoogleDrive:
		return parseGoogleDriveURL(rawURL)
	default:
		return ParsedURL{}, ErrUnsupportedSourceType
	}
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

	return ParsedURL{
		SourceType: constants.TeacherIntroVideoSourceYouTube,
		URL:        fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
	}, nil
}

func parseGoogleDriveURL(rawURL string) (ParsedURL, error) {
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

	host := strings.ToLower(parsed.Host)
	if host != "drive.google.com" && host != "docs.google.com" {
		return ParsedURL{}, ErrInvalidURL
	}

	if strings.Contains(strings.ToLower(trimmed), "/folders/") {
		return ParsedURL{}, ErrInvalidURL
	}

	if matches := googleDriveFilePattern.FindStringSubmatch(trimmed); len(matches) == 2 {
		fileID := matches[1]
		return ParsedURL{
			SourceType: constants.TeacherIntroVideoSourceGoogleDrive,
			URL:        fmt.Sprintf("https://drive.google.com/file/d/%s/view", fileID),
		}, nil
	}

	fileID := parsed.Query().Get("id")
	if fileID == "" {
		return ParsedURL{}, ErrInvalidURL
	}

	return ParsedURL{
		SourceType: constants.TeacherIntroVideoSourceGoogleDrive,
		URL:        fmt.Sprintf("https://drive.google.com/file/d/%s/view", fileID),
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
