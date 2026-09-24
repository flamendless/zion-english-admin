package constants

type FeatureFlagKey string

const (
	FeatureFlagIntegrationZoom           FeatureFlagKey = "integration.zoom"
	FeatureFlagIntegrationGoogleCalendar FeatureFlagKey = "integration.google_calendar"
	FeatureFlagClassOverdueGracePeriod     FeatureFlagKey = "class.overdue_grace_period_minutes"
	FeatureFlagIntroVideoUploads         FeatureFlagKey = "intro_video.uploads"
	FeatureFlagIntroVideoCompressPreset  FeatureFlagKey = "intro_video.compress_preset"
	FeatureFlagPersistentOnboarding    FeatureFlagKey = "onboarding.persistent"
)

const (
	DefaultClassOverdueGracePeriodMinutes = 0
	MaxClassOverdueGracePeriodMinutes     = 1440
)
