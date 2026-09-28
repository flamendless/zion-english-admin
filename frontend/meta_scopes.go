package frontend

import "zion-english/internal/constants"

func MetaTagScopesSiteWide() []constants.MetaTagScope {
	return []constants.MetaTagScope{constants.MetaTagScopeSiteWide}
}

func MetaTagScopesLanding() []constants.MetaTagScope {
	return []constants.MetaTagScope{constants.MetaTagScopeSiteWide, constants.MetaTagScopePublicPages}
}

func MetaTagScopesLoginPublic() []constants.MetaTagScope {
	return []constants.MetaTagScope{constants.MetaTagScopePublicPages}
}
