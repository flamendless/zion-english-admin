package constants

type IntroVideoCompressPreset string

const (
	IntroVideoCompressPresetVeryLow  IntroVideoCompressPreset = "very_low"
	IntroVideoCompressPresetMedium   IntroVideoCompressPreset = "medium"
	IntroVideoCompressPresetFine     IntroVideoCompressPreset = "fine"
	IntroVideoCompressPresetHigh     IntroVideoCompressPreset = "high"
	IntroVideoCompressPresetOriginal IntroVideoCompressPreset = "original"
)

const (
	introVideoCompressTimeoutDefaultSeconds = 300
	introVideoCompressTimeoutHighSeconds    = 600
	// Extra encode budget per MB above IntroVideoCompressMinBytes.
	introVideoCompressTimeoutSecondsPerMB = 15
	introVideoCompressTimeoutMaxSeconds   = 3600
)

type IntroVideoEncodeSettings struct {
	SkipCompress bool
	CRF          string
	Preset       string
	MaxWidth     int
	AudioBitrate string
	TimeoutSecs  int
}

var introVideoCompressPresetOptions = []IntroVideoCompressPreset{
	IntroVideoCompressPresetVeryLow,
	IntroVideoCompressPresetMedium,
	IntroVideoCompressPresetFine,
	IntroVideoCompressPresetHigh,
	IntroVideoCompressPresetOriginal,
}

func DefaultIntroVideoCompressPreset() IntroVideoCompressPreset {
	return IntroVideoCompressPresetFine
}

func IntroVideoCompressPresetOptions() []IntroVideoCompressPreset {
	return introVideoCompressPresetOptions
}

func ValidIntroVideoCompressPreset(value string) bool {
	switch IntroVideoCompressPreset(value) {
	case IntroVideoCompressPresetVeryLow,
		IntroVideoCompressPresetMedium,
		IntroVideoCompressPresetFine,
		IntroVideoCompressPresetHigh,
		IntroVideoCompressPresetOriginal:
		return true
	default:
		return false
	}
}

func IntroVideoCompressPresetLabel(preset IntroVideoCompressPreset) string {
	switch preset {
	case IntroVideoCompressPresetVeryLow:
		return "Very small size (very low quality)"
	case IntroVideoCompressPresetMedium:
		return "Medium"
	case IntroVideoCompressPresetFine:
		return "Fine"
	case IntroVideoCompressPresetHigh:
		return "High quality"
	case IntroVideoCompressPresetOriginal:
		return "Original (no re-encode)"
	default:
		return string(preset)
	}
}

func IntroVideoEncodeSettingsForPreset(preset IntroVideoCompressPreset) IntroVideoEncodeSettings {
	switch preset {
	case IntroVideoCompressPresetVeryLow:
		return IntroVideoEncodeSettings{
			CRF:          "32",
			Preset:       "fast",
			MaxWidth:     854,
			AudioBitrate: "96k",
			TimeoutSecs:  introVideoCompressTimeoutDefaultSeconds,
		}
	case IntroVideoCompressPresetMedium:
		return IntroVideoEncodeSettings{
			CRF:          "28",
			Preset:       "medium",
			MaxWidth:     1280,
			AudioBitrate: "128k",
			TimeoutSecs:  introVideoCompressTimeoutDefaultSeconds,
		}
	case IntroVideoCompressPresetHigh:
		return IntroVideoEncodeSettings{
			CRF:          "18",
			Preset:       "slow",
			MaxWidth:     1920,
			AudioBitrate: "192k",
			TimeoutSecs:  introVideoCompressTimeoutHighSeconds,
		}
	case IntroVideoCompressPresetOriginal:
		return IntroVideoEncodeSettings{
			SkipCompress: true,
		}
	case IntroVideoCompressPresetFine:
		fallthrough
	default:
		return IntroVideoEncodeSettings{
			CRF:          "22",
			Preset:       "medium",
			MaxWidth:     1920,
			AudioBitrate: "160k",
			TimeoutSecs:  introVideoCompressTimeoutDefaultSeconds,
		}
	}
}

func IntroVideoEncodeSettingsForPresetValue(value string) IntroVideoEncodeSettings {
	if !ValidIntroVideoCompressPreset(value) {
		return IntroVideoEncodeSettingsForPreset(DefaultIntroVideoCompressPreset())
	}
	return IntroVideoEncodeSettingsForPreset(IntroVideoCompressPreset(value))
}

func IntroVideoShouldCompress(fileSize int64, settings IntroVideoEncodeSettings) bool {
	if settings.SkipCompress || fileSize < IntroVideoCompressMinBytes {
		return false
	}
	return true
}

// IntroVideoCompressTimeoutSecs returns the ffmpeg time limit for a staged upload.
// Large files get extra time above the preset base timeout, capped at introVideoCompressTimeoutMaxSeconds.
func IntroVideoCompressTimeoutSecs(fileSize int64, settings IntroVideoEncodeSettings) int {
	base := settings.TimeoutSecs
	if base <= 0 {
		base = introVideoCompressTimeoutDefaultSeconds
	}
	if fileSize <= IntroVideoCompressMinBytes {
		return base
	}
	extraBytes := fileSize - IntroVideoCompressMinBytes
	extraMB := (extraBytes + (1<<20) - 1) / (1 << 20)
	scaled := base + int(extraMB)*introVideoCompressTimeoutSecondsPerMB
	if scaled > introVideoCompressTimeoutMaxSeconds {
		return introVideoCompressTimeoutMaxSeconds
	}
	return scaled
}
