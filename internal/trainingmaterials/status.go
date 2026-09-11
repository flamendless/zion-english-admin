package trainingmaterials

type Status string

const (
	StatusPublished Status = "published"
	StatusDraft     Status = "draft"
	StatusDeleted   Status = "deleted"
)

func ValidStatus(status string) bool {
	switch Status(status) {
	case StatusPublished, StatusDraft, StatusDeleted:
		return true
	default:
		return false
	}
}

func ValidFormStatus(status string) bool {
	switch Status(status) {
	case StatusPublished, StatusDraft:
		return true
	default:
		return false
	}
}
