package constants

type MetaTagAttr string

const (
	MetaTagAttrName     MetaTagAttr = "name"
	MetaTagAttrProperty MetaTagAttr = "property"
)

type MetaTagScope string

const (
	MetaTagScopeSiteWide    MetaTagScope = "site_wide"
	MetaTagScopePublicPages MetaTagScope = "public_pages"
)

func ValidMetaTagAttr(attr string) bool {
	switch MetaTagAttr(attr) {
	case MetaTagAttrName, MetaTagAttrProperty:
		return true
	default:
		return false
	}
}

func ValidMetaTagScope(scope string) bool {
	switch MetaTagScope(scope) {
	case MetaTagScopeSiteWide, MetaTagScopePublicPages:
		return true
	default:
		return false
	}
}

func MetaTagScopeLabel(scope MetaTagScope) string {
	switch scope {
	case MetaTagScopeSiteWide:
		return "Site-wide"
	case MetaTagScopePublicPages:
		return "Public pages"
	default:
		return string(scope)
	}
}

func MetaTagAttrLabel(attr MetaTagAttr) string {
	switch attr {
	case MetaTagAttrName:
		return "name"
	case MetaTagAttrProperty:
		return "property"
	default:
		return string(attr)
	}
}
