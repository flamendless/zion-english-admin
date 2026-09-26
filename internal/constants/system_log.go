package constants

type SystemLogUploadOutcome string

const (
	SystemLogUploadOutcomeFailed    SystemLogUploadOutcome = "failed"
	SystemLogUploadOutcomeSucceeded SystemLogUploadOutcome = "succeeded"
)

const SystemLogUploadLogPrefix = "upload log: "

// Legacy rows written before upload logs used a single error prefix.
const SystemLogUploadLegacyErrorPrefix = "upload error: "
