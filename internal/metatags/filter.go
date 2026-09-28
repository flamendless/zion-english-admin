package metatags

import (
	"zion-english/internal/constants"
)

func FilterTagsByScopeList(tags []Tag, scopes []constants.MetaTagScope) []Tag {
	return FilterTags(tags, scopes...)
}

func FilterTags(tags []Tag, scopes ...constants.MetaTagScope) []Tag {
	if len(scopes) == 0 {
		return tags
	}
	allowed := make(map[constants.MetaTagScope]bool, len(scopes))
	for _, scope := range scopes {
		allowed[scope] = true
	}
	out := make([]Tag, 0, len(tags))
	for _, tag := range tags {
		if allowed[tag.Scope] {
			out = append(out, tag)
		}
	}
	return out
}

func OgTitle(tags []Tag) string {
	for _, tag := range tags {
		if tag.Attr == constants.MetaTagAttrProperty && tag.Name == "og:title" && tag.Content != "" {
			return tag.Content
		}
	}
	return ""
}
