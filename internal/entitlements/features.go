package entitlements

type Feature string

const (
	FeatureAnalyticsPage         Feature = "analytics.page"
	FeatureAnalyticsDateRange    Feature = "analytics.date_range"
	FeatureAnalyticsSummary      Feature = "analytics.summary"
	FeatureAnalyticsTrial        Feature = "analytics.trial"
	FeatureAnalyticsChartsWeekly Feature = "analytics.charts_weekly"
	FeatureAnalyticsByStudent    Feature = "analytics.by_student"
	FeatureAnalyticsNoShows      Feature = "analytics.no_shows"
	FeatureAnalyticsRetention    Feature = "analytics.retention"
	FeatureExportPage            Feature = "export.page"
	FeatureExportClasses         Feature = "export.classes"
	FeatureExportStudents        Feature = "export.students"
	FeatureExportSchedule        Feature = "export.schedule"
	FeatureNotificationsEmail    Feature = "notifications.email"
)
