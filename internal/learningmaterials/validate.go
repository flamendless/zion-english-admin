package learningmaterials

import (
	"errors"
	"net/url"
	"strings"
)

const (
	MinTags     = 1
	MaxTags     = 7
	MaxTitleLen = 64
)

var (
	ErrTitleRequired       = errors.New("title is required")
	ErrTitleTooLong        = errors.New("title must be 64 characters or fewer")
	ErrDescriptionRequired = errors.New("description is required")
	ErrURLRequired         = errors.New("url is required")
	ErrInvalidURL          = errors.New("url must be a valid http or https link")
	ErrInvalidAccess       = errors.New("access must be public or private")
	ErrInvalidStatus       = errors.New("status must be published, draft, or deleted")
	ErrTagCount            = errors.New("each material must have between 1 and 7 tags")
	ErrTagLabelRequired    = errors.New("tag labels cannot be empty")
	ErrTagLabelTooLong     = errors.New("tag labels must be 40 characters or fewer")
)

type Request struct {
	Title        string
	Description  string
	URL          string
	ThumbnailURL string
	Access       string
	Status       string
	TagLabels    []string
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

func ValidateEditRequest(req Request) error {
	if err := validateRequestFields(req); err != nil {
		return err
	}
	if !ValidStatus(req.Status) {
		return ErrInvalidStatus
	}
	return nil
}

func ValidateRequest(req Request) error {
	if err := validateRequestFields(req); err != nil {
		return err
	}
	if !ValidFormStatus(req.Status) {
		return ErrInvalidStatus
	}
	return nil
}

func validateRequestFields(req Request) error {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return ErrTitleRequired
	}
	if len(title) > MaxTitleLen {
		return ErrTitleTooLong
	}
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
