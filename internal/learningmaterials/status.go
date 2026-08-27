package learningmaterials

const (
	StatusPublished = "published"
	StatusDraft     = "draft"
	StatusDeleted   = "deleted"
)

func ValidStatus(status string) bool {
	switch status {
	case StatusPublished, StatusDraft, StatusDeleted:
		return true
	default:
		return false
	}
}

func ValidFormStatus(status string) bool {
	switch status {
	case StatusPublished, StatusDraft:
		return true
	default:
		return false
	}
}
