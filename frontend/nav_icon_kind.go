package frontend

type NavIconKind string

const (
	NavIconProfile          NavIconKind = "profile"
	NavIconGuides           NavIconKind = "guides"
	NavIconLibrary          NavIconKind = "library"
	NavIconDocuments        NavIconKind = "documents"
	NavIconTeachers         NavIconKind = "teachers"
	NavIconStudents         NavIconKind = "students"
	NavIconMyStudents       NavIconKind = "my-students"
	NavIconClasses          NavIconKind = "classes"
	NavIconSchedule         NavIconKind = "schedule"
	NavIconScheduleSeries   NavIconKind = "schedule-series"
	NavIconReports          NavIconKind = "reports"
	NavIconAnalytics        NavIconKind = "analytics"
	NavIconProcess          NavIconKind = "process"
	NavIconFeatureFlags     NavIconKind = "feature-flags"
	NavIconLogs             NavIconKind = "logs"
	NavIconAnnouncements    NavIconKind = "announcements"
	NavIconChangelogs       NavIconKind = "changelogs"
	NavIconPeopleGroup      NavIconKind = "people-group"
	NavIconResourcesGroup   NavIconKind = "resources-group"
	NavIconInsightsGroup    NavIconKind = "insights-group"
	NavIconAdminGroup       NavIconKind = "admin-group"
	NavIconDefault          NavIconKind = "default"
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
	"/profile":            NavIconProfile,
	"/guides":             NavIconGuides,
	"/learning-materials":   NavIconLibrary,
	"/documents":            NavIconDocuments,
	"/teachers":             NavIconTeachers,
	"/students":             NavIconStudents,
	"/classes":              NavIconClasses,
	"/schedule":             NavIconSchedule,
	"/schedule/series":      NavIconScheduleSeries,
	"/my-students":          NavIconMyStudents,
	"/reports":              NavIconReports,
	"/analytics":            NavIconAnalytics,
	"/student-relationships": NavIconStudents,
	"/process":              NavIconProcess,
	"/feature-flags":        NavIconFeatureFlags,
	"/logs":                 NavIconLogs,
	"/announcements":        NavIconAnnouncements,
	"/changelogs":           NavIconChangelogs,
}

var navIconByGroupID = map[string]NavIconKind{
	"classes":   NavIconClasses,
	"people":    NavIconPeopleGroup,
	"resources": NavIconResourcesGroup,
	"insights":  NavIconInsightsGroup,
	"admin":     NavIconAdminGroup,
}

var navIconToneByKind = map[NavIconKind]NavIconTone{
	NavIconProfile:        NavIconTonePrimary,
	NavIconGuides:         NavIconTonePrimary,
	NavIconLibrary:        NavIconTonePrimary,
	NavIconDocuments:      NavIconToneInfo,
	NavIconTeachers:       NavIconToneSuccess,
	NavIconStudents:       NavIconToneSuccess,
	NavIconMyStudents:     NavIconToneSuccess,
	NavIconClasses:        NavIconTonePrimary,
	NavIconSchedule:       NavIconTonePrimary,
	NavIconScheduleSeries: NavIconToneInfo,
	NavIconReports:        NavIconToneInfo,
	NavIconAnalytics:      NavIconToneInfo,
	NavIconProcess:        NavIconToneWarning,
	NavIconFeatureFlags:   NavIconToneWarning,
	NavIconLogs:           NavIconToneNeutral,
	NavIconAnnouncements:  NavIconToneWarning,
	NavIconChangelogs:     NavIconToneNeutral,
	NavIconPeopleGroup:    NavIconToneSuccess,
	NavIconResourcesGroup: NavIconTonePrimary,
	NavIconInsightsGroup:  NavIconToneInfo,
	NavIconAdminGroup:     NavIconToneWarning,
	NavIconDefault:        NavIconToneNeutral,
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
