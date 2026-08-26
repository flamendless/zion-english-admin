package frontend

type IconKind string

const (
	IconKindView       IconKind = "view"
	IconKindEdit       IconKind = "edit"
	IconKindConduct    IconKind = "conduct"
	IconKindApprove    IconKind = "approve"
	IconKindCancel     IconKind = "cancel"
	IconKindDelete     IconKind = "delete"
	IconKindWarning    IconKind = "warning"
	IconKindReject     IconKind = "reject"
	IconKindUnapprove  IconKind = "unapprove"
	IconKindExternal   IconKind = "external"
	IconKindDownload   IconKind = "download"
	IconKindGenerate   IconKind = "generate"
	IconKindRegenerate IconKind = "regenerate"
	IconKindMarkRead   IconKind = "mark-read"
)

func iconActionClass(kind IconKind) string {
	return "icon-action-btn icon-action-btn--" + string(kind)
}
