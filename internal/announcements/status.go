package announcements

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

func AllStatuses() []string {
	return []string{StatusPublished, StatusDraft, StatusDeleted}
}

func FormStatuses() []string {
	return []string{StatusPublished, StatusDraft}
}
