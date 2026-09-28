package metatags

import (
	"net/url"
	"strings"
	"unicode/utf8"

	"zion-english/internal/constants"
	"zion-english/internal/utils"
)

const maxNameLen = 120
const maxAttrLen = 2048

type Request struct {
	Name      string
	Content   string
	Value     string
	SortOrder int64
	Attr      constants.MetaTagAttr
	Scope     constants.MetaTagScope
}

func NormalizeRequest(req Request) Request {
	req.Name = strings.TrimSpace(req.Name)
	req.Content = strings.TrimSpace(req.Content)
	req.Value = strings.TrimSpace(req.Value)
	if !constants.ValidMetaTagAttr(string(req.Attr)) {
		req.Attr = constants.MetaTagAttrName
	}
	if !constants.ValidMetaTagScope(string(req.Scope)) {
		req.Scope = constants.MetaTagScopeSiteWide
	}
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
	if req.Attr == constants.MetaTagAttrName && isReservedMetaName(req.Name) {
		return ErrReservedName
	}
	if !constants.ValidMetaTagAttr(string(req.Attr)) {
		return ErrInvalidAttr
	}
	if !constants.ValidMetaTagScope(string(req.Scope)) {
		return ErrInvalidScope
	}
	if req.Attr == constants.MetaTagAttrProperty && strings.EqualFold(req.Name, "og:image") && req.Content != "" && !isHTTPURL(req.Content) {
		return ErrInvalidOgImageURL
	}
	if req.Content == "" && req.Value == "" {
		return ErrContentOrValueRequired
	}
	if utf8.RuneCountInString(req.Content) > maxAttrLen || utf8.RuneCountInString(req.Value) > maxAttrLen {
		return ErrAttributeTooLong
	}
	return nil
}

func isHTTPURL(raw string) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func isReservedMetaName(name string) bool {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "title", "viewport":
		return true
	default:
		return false
	}
}
