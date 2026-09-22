package trainingmaterials

import (
	"strings"
)

const (
	MinTags     = 1
	MaxTags     = 7
	MaxTitleLen = 64
)

type Request struct {
	Title       string
	Description string
	URL         string
	SourceType  string
	Status      string
	Required    bool
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

	sourceType := strings.TrimSpace(req.SourceType)
	if sourceType == "" {
		inferred, err := InferSourceType(req.URL)
		if err != nil {
			return err
		}
		sourceType = string(inferred)
	}
	if !ValidSourceType(sourceType) {
		return ErrUnsupportedSourceType
	}
	if _, err := ParseURL(SourceType(sourceType), req.URL); err != nil {
		return err
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
