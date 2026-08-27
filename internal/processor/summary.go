package processor

import (
	"zion-english/internal/constants"
	"zion-english/internal/utils"
)

type RecordSummary struct {
	TotalClasses    int
	TotalConducted  int
	TotalCancelled  int
	TotalDuration   int
	DurationDisplay string
}

func ComputeRecordSummary(records []ClassRecord) RecordSummary {
	var summary RecordSummary
	for _, rec := range records {
		summary.TotalClasses++
		switch constants.ClassStatus(rec.Status) {
		case constants.ClassStatusConducted:
			summary.TotalConducted++
		case constants.ClassStatusCancelled:
			summary.TotalCancelled++
		}
		summary.TotalDuration += rec.DurationInMinutes
	}
	summary.DurationDisplay = utils.FormatDurationMinutes(int64(summary.TotalDuration))
	return summary
}
