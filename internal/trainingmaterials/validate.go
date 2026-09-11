package trainingmaterials

import (
	"errors"
	"net/url"
	"strings"
)

var (
	ErrTitleRequired = errors.New("title is required")
	ErrURLRequired   = errors.New("url is required")
	ErrInvalidURL    = errors.New("url must be a valid YouTube or Google Drive link")
	ErrInvalidStatus = errors.New("status must be published or draft")
)

type Request struct {
	Title       string
	Description string
	URL         string
	Status      string
}

func ValidateRequest(req Request) error {
	if strings.TrimSpace(req.Title) == "" {
		return ErrTitleRequired
	}
	if strings.TrimSpace(req.URL) == "" {
		return ErrURLRequired
	}
	if !IsAllowedURL(req.URL) {
		return ErrInvalidURL
	}
	if !ValidFormStatus(req.Status) {
		return ErrInvalidStatus
	}
	return nil
}

func IsAllowedURL(raw string) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}

	host := strings.ToLower(parsed.Hostname())
	switch host {
	case "youtu.be", "drive.google.com", "docs.google.com":
		return true
	}
	if strings.HasSuffix(host, ".youtube.com") || strings.HasSuffix(host, ".youtube-nocookie.com") {
		return true
	}
	return false
}
