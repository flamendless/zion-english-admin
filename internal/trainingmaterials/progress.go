package trainingmaterials

const CompletionThreshold = 90.0

func ComputeProgress(watchSeconds int64, durationSeconds int64) float64 {
	if durationSeconds <= 0 {
		return 0
	}
	percent := float64(watchSeconds) / float64(durationSeconds) * 100
	if percent > 100 {
		return 100
	}
	return percent
}

func IsCompleted(progressPercent float64) bool {
	return progressPercent >= CompletionThreshold
}
