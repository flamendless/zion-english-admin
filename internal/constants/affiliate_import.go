package constants

type AffiliateImportDiffStatus string

const (
	AffiliateImportDiffNew            AffiliateImportDiffStatus = "new"
	AffiliateImportDiffChanged        AffiliateImportDiffStatus = "changed"
	AffiliateImportDiffUnchanged      AffiliateImportDiffStatus = "unchanged"
	AffiliateImportDiffDuplicateInCSV AffiliateImportDiffStatus = "duplicate_in_csv"
	AffiliateImportDiffMissingItemKey AffiliateImportDiffStatus = "missing_item_key"
)

func AffiliateImportDiffStatusLabel(status AffiliateImportDiffStatus) string {
	switch status {
	case AffiliateImportDiffNew:
		return "New"
	case AffiliateImportDiffChanged:
		return "Changed"
	case AffiliateImportDiffUnchanged:
		return "Unchanged"
	case AffiliateImportDiffDuplicateInCSV:
		return "Duplicate in CSV"
	case AffiliateImportDiffMissingItemKey:
		return "Missing item key"
	default:
		return string(status)
	}
}
