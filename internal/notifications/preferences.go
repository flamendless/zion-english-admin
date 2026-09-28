package notifications

import (
	"context"
	"zion-english/internal/database/queries"
)

func PreferenceMap(rows []queries.GetTeacherNotificationPreferencesRow) map[Category]bool {
	out := make(map[Category]bool, len(rows))
	for _, row := range rows {
		out[Category(row.Category)] = row.Enabled != 0
	}
	return out
}

func CategoryEnabled(prefs map[Category]bool, category Category) bool {
	if prefs == nil {
		return true
	}
	if enabled, ok := prefs[category]; ok {
		return enabled
	}
	return true
}

func (s *Service) teacherAcceptsKind(ctx context.Context, teacherID int64, kind string) bool {
	category, ok := CategoryForKind(kind)
	if !ok {
		return true
	}
	rows, err := s.q.GetTeacherNotificationPreferences(ctx, teacherID)
	if err != nil {
		return true
	}
	return CategoryEnabled(PreferenceMap(rows), category)
}
