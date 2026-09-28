package frontend

type NavIconKind string

const (
	NavIconProfile             NavIconKind = "profile"
	NavIconGuides              NavIconKind = "guides"
	NavIconLibrary             NavIconKind = "library"
	NavIconTraining            NavIconKind = "training"
	NavIconDocuments           NavIconKind = "documents"
	NavIconIntroVideos         NavIconKind = "intro-videos"
	NavIconTeachers            NavIconKind = "teachers"
	NavIconStudents            NavIconKind = "students"
	NavIconMyStudents          NavIconKind = "my-students"
	NavIconClasses             NavIconKind = "classes"
	NavIconSchedule            NavIconKind = "schedule"
	NavIconScheduleSeries      NavIconKind = "schedule-series"
	NavIconRecordClass         NavIconKind = "record-class"
	NavIconReports             NavIconKind = "reports"
	NavIconReportsHistory      NavIconKind = "reports-history"
	NavIconAnalytics           NavIconKind = "analytics"
	NavIconAffiliateAdsReports NavIconKind = "affiliate-ads-reports"
	NavIconPayments            NavIconKind = "payments"
	NavIconProcess             NavIconKind = "process"
	NavIconFeatureFlags        NavIconKind = "feature-flags"
	NavIconSettings            NavIconKind = "settings"
	NavIconMeta                NavIconKind = "meta"
	NavIconAffiliates          NavIconKind = "affiliates"
	NavIconAds                 NavIconKind = "ads"
	NavIconLogs                NavIconKind = "logs"
	NavIconUploadLogs          NavIconKind = "upload-logs"
	NavIconAnnouncements       NavIconKind = "announcements"
	NavIconChangelogs          NavIconKind = "changelogs"
	NavIconPeopleGroup         NavIconKind = "people-group"
	NavIconResourcesGroup      NavIconKind = "resources-group"
	NavIconInsightsGroup       NavIconKind = "insights-group"
	NavIconAdminGroup          NavIconKind = "admin-group"
	NavIconDefault             NavIconKind = "default"
)

type NavIconTone string

const (
	NavIconTonePrimary NavIconTone = "primary"
	NavIconToneInfo    NavIconTone = "info"
	NavIconToneSuccess NavIconTone = "success"
	NavIconToneWarning NavIconTone = "warning"
	NavIconToneNeutral NavIconTone = "neutral"
)

var navIconByPath = map[string]NavIconKind{
	"/profile":               NavIconProfile,
	"/guides":                NavIconGuides,
	"/learning-materials":    NavIconLibrary,
	"/training-materials":    NavIconTraining,
	"/documents":             NavIconDocuments,
	"/intro-videos":          NavIconIntroVideos,
	"/teachers":              NavIconTeachers,
	"/students":              NavIconStudents,
	"/classes":               NavIconClasses,
	"/schedule":              NavIconSchedule,
	"/schedule/series":       NavIconScheduleSeries,
	"/my-students":           NavIconMyStudents,
	"/reports":               NavIconReports,
	"/reports/history":       NavIconReportsHistory,
	"/affiliate-ads/reports": NavIconAffiliateAdsReports,
	"/analytics":             NavIconAnalytics,
	"/payments":              NavIconPayments,
	"/student-relationships": NavIconStudents,
	"/process":               NavIconProcess,
	"/feature-flags":         NavIconFeatureFlags,
	"/settings":              NavIconSettings,
	"/meta":                  NavIconMeta,
	"/affiliates":            NavIconAffiliates,
	"/ads":                   NavIconAds,
	"/logs":                  NavIconLogs,
	"/upload-logs":           NavIconUploadLogs,
	"/announcements":         NavIconAnnouncements,
	"/changelogs":            NavIconChangelogs,
}

var navIconByGroupID = map[string]NavIconKind{
	"classes":   NavIconClasses,
	"people":    NavIconPeopleGroup,
	"resources": NavIconResourcesGroup,
	"insights":  NavIconInsightsGroup,
	"admin":     NavIconAdminGroup,
}

var navIconToneByKind = map[NavIconKind]NavIconTone{
	NavIconProfile:             NavIconTonePrimary,
	NavIconGuides:              NavIconTonePrimary,
	NavIconLibrary:             NavIconTonePrimary,
	NavIconTraining:            NavIconTonePrimary,
	NavIconDocuments:           NavIconToneInfo,
	NavIconIntroVideos:         NavIconToneInfo,
	NavIconTeachers:            NavIconToneSuccess,
	NavIconStudents:            NavIconToneSuccess,
	NavIconMyStudents:          NavIconToneSuccess,
	NavIconClasses:             NavIconTonePrimary,
	NavIconSchedule:            NavIconTonePrimary,
	NavIconScheduleSeries:      NavIconToneInfo,
	NavIconRecordClass:         NavIconToneSuccess,
	NavIconReports:             NavIconToneInfo,
	NavIconReportsHistory:      NavIconToneInfo,
	NavIconAnalytics:           NavIconToneInfo,
	NavIconAffiliateAdsReports: NavIconToneInfo,
	NavIconPayments:            NavIconToneSuccess,
	NavIconProcess:             NavIconToneWarning,
	NavIconFeatureFlags:        NavIconToneWarning,
	NavIconSettings:            NavIconToneWarning,
	NavIconMeta:                NavIconToneWarning,
	NavIconAffiliates:          NavIconToneWarning,
	NavIconAds:                 NavIconToneWarning,
	NavIconLogs:                NavIconToneNeutral,
	NavIconUploadLogs:          NavIconToneInfo,
	NavIconAnnouncements:       NavIconToneWarning,
	NavIconChangelogs:          NavIconToneNeutral,
	NavIconPeopleGroup:         NavIconToneSuccess,
	NavIconResourcesGroup:      NavIconTonePrimary,
	NavIconInsightsGroup:       NavIconToneInfo,
	NavIconAdminGroup:          NavIconToneWarning,
	NavIconDefault:             NavIconToneNeutral,
}

func NavIconForPath(path string) NavIconKind {
	if icon, ok := navIconByPath[path]; ok {
		return icon
	}
	return NavIconDefault
}

func NavIconForGroupID(groupID string) NavIconKind {
	if icon, ok := navIconByGroupID[groupID]; ok {
		return icon
	}
	return NavIconDefault
}

func NavIconToneForKind(kind NavIconKind) NavIconTone {
	if tone, ok := navIconToneByKind[kind]; ok {
		return tone
	}
	return NavIconToneNeutral
}

func navIconBadgeClass(kind NavIconKind) string {
	return "nav-icon-badge nav-icon-badge--" + string(NavIconToneForKind(kind))
}

func navChoiceIconClass(kind NavIconKind) string {
	return "choice-option-card__icon choice-option-card__icon--" + string(NavIconToneForKind(kind))
}
