package constants

type UploadLogOutcome string

const (
	UploadLogOutcomeFailed    UploadLogOutcome = "failed"
	UploadLogOutcomeSucceeded UploadLogOutcome = "succeeded"
)

type UploadLogKind string

const (
	UploadLogKindIntroVideo UploadLogKind = "intro_video"
	UploadLogKindAvatar     UploadLogKind = "avatar"
	UploadLogKindDocument   UploadLogKind = "document"
	UploadLogKindOther      UploadLogKind = "other"
)

func (k UploadLogKind) Label() string {
	switch k {
	case UploadLogKindIntroVideo:
		return "Intro video"
	case UploadLogKindAvatar:
		return "Avatar"
	case UploadLogKindDocument:
		return "Document"
	default:
		return "Other"
	}
}

func ValidUploadLogKind(value string) bool {
	switch UploadLogKind(value) {
	case UploadLogKindIntroVideo, UploadLogKindAvatar, UploadLogKindDocument, UploadLogKindOther:
		return true
	default:
		return false
	}
}
