package learningmaterials

const (
	AccessPublic  = "public"
	AccessPrivate = "private"
)

func ValidAccess(access string) bool {
	switch access {
	case AccessPublic, AccessPrivate:
		return true
	default:
		return false
	}
}
