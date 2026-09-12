package constants

type FeatureFlagKey string

const (
	FeatureFlagIntegrationZoom           FeatureFlagKey = "integration.zoom"
	FeatureFlagIntegrationGoogleCalendar FeatureFlagKey = "integration.google_calendar"
	FeatureFlagClassOverdueGracePeriod    FeatureFlagKey = "class.overdue_grace_period_minutes"
)

const (
	DefaultClassOverdueGracePeriodMinutes = 0
	MaxClassOverdueGracePeriodMinutes     = 1440
)
