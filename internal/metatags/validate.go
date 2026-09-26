package metatags

import (
	"strings"
	"unicode/utf8"

	"zion-english/internal/utils"
)

const maxNameLen = 120
const maxAttrLen = 2048

type Request struct {
	Name      string
	Content   string
	Value     string
	SortOrder int64
}

func NormalizeRequest(req Request) Request {
	req.Name = strings.TrimSpace(req.Name)
	req.Content = strings.TrimSpace(req.Content)
	req.Value = strings.TrimSpace(req.Value)
	return req
}

func ValidateRequest(req Request) error {
	req = NormalizeRequest(req)
	if utils.IsBlank(req.Name) {
		return ErrNameRequired
	}
	if utf8.RuneCountInString(req.Name) > maxNameLen {
		return ErrInvalidName
	}
	if req.Content == "" && req.Value == "" {
		return ErrContentOrValueRequired
	}
	if utf8.RuneCountInString(req.Content) > maxAttrLen || utf8.RuneCountInString(req.Value) > maxAttrLen {
		return ErrAttributeTooLong
	}
	return nil
}
