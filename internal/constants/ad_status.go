package constants

type AdStatus string

const (
	AdStatusPublished AdStatus = "published"
	AdStatusDraft     AdStatus = "draft"
	AdStatusDeleted   AdStatus = "deleted"
)

func ValidAdStatus(status string) bool {
	switch AdStatus(status) {
	case AdStatusPublished, AdStatusDraft, AdStatusDeleted:
		return true
	default:
		return false
	}
}

func ValidAdFormStatus(status string) bool {
	switch AdStatus(status) {
	case AdStatusPublished, AdStatusDraft:
		return true
	default:
		return false
	}
}
