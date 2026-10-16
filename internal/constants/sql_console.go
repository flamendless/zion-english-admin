package constants

const (
	SQLConsoleMaxQueryBytes     = 16 * 1024
	SQLConsoleMaxSavedNameLen   = 120
	SQLConsoleDefaultRowLimit   = 500
	SQLConsoleQueryTimeoutSec   = 30
	SQLConsoleHistoryPageSize   = 25
	SQLConsoleHistoryPreviewLen = 80
)

type SQLConsoleQueryOutcome string

const (
	SQLConsoleOutcomeSucceeded SQLConsoleQueryOutcome = "succeeded"
	SQLConsoleOutcomeFailed    SQLConsoleQueryOutcome = "failed"
	SQLConsoleOutcomeRejected  SQLConsoleQueryOutcome = "rejected"
)
