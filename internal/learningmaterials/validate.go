package learningmaterials

import (
	"errors"
	"net/url"
	"strings"
)

const (
	MinTags = 1
	MaxTags = 7
)

var (
	ErrDescriptionRequired = errors.New("description is required")
	ErrURLRequired         = errors.New("url is required")
	ErrInvalidURL          = errors.New("url must be a valid http or https link")
	ErrInvalidAccess       = errors.New("access must be public or private")
	ErrInvalidStatus       = errors.New("status must be published or draft")
	ErrTagCount            = errors.New("each material must have between 1 and 7 tags")
	ErrTagLabelRequired    = errors.New("tag labels cannot be empty")
	ErrTagLabelTooLong     = errors.New("tag labels must be 40 characters or fewer")
)

type Request struct {
	Description string
	URL         string
	Access      string
	Status      string
	TagLabels   []string
}

func NormalizeTagLabels(labels []string) []string {
	seen := make(map[string]bool, len(labels))
	out := make([]string, 0, len(labels))
	for _, raw := range labels {
		label := strings.ToLower(strings.TrimSpace(raw))
		if label == "" || seen[label] {
			continue
		}
		seen[label] = true
		out = append(out, label)
	}
	return out
}

func ValidateRequest(req Request) error {
	if strings.TrimSpace(req.Description) == "" {
		return ErrDescriptionRequired
	}
	if strings.TrimSpace(req.URL) == "" {
		return ErrURLRequired
	}
	if !isValidURL(req.URL) {
		return ErrInvalidURL
	}
	if !ValidAccess(req.Access) {
		return ErrInvalidAccess
	}
	if !ValidFormStatus(req.Status) {
		return ErrInvalidStatus
	}

	labels := NormalizeTagLabels(req.TagLabels)
	if len(labels) < MinTags || len(labels) > MaxTags {
		return ErrTagCount
	}
	for _, label := range labels {
		if len(label) > 40 {
			return ErrTagLabelTooLong
		}
	}
	return nil
}

func isValidURL(raw string) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}
