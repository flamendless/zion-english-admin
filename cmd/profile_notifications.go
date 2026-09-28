package cmd

import (
	"net/http"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/notifications"

	"go.uber.org/zap"
)

func buildNotificationPreferenceItems(rows []queries.GetTeacherNotificationPreferencesRow) []frontend.NotificationPreferenceItem {
	prefs := notifications.PreferenceMap(rows)
	items := make([]frontend.NotificationPreferenceItem, 0, len(notifications.TeacherNotificationCategories()))
	for _, def := range notifications.TeacherNotificationCategories() {
		items = append(items, frontend.NotificationPreferenceItem{
			Category:    string(def.Category),
			Label:       def.Label,
			Description: def.Description,
			Enabled:     notifications.CategoryEnabled(prefs, def.Category),
		})
	}
	return items
}

func notificationPrefFormKey(category notifications.Category) string {
	return "pref_" + string(category)
}

func handleProfileNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)

	if err := r.ParseForm(); err != nil {
		setErrorFlash(w, "Invalid form submission")
		HttpRedirect(w, r, "/profile?tab=notifications")
		return
	}

	q := dbRW.GetQueries()
	for _, def := range notifications.TeacherNotificationCategories() {
		enabled := int64(0)
		if r.FormValue(notificationPrefFormKey(def.Category)) == "1" {
			enabled = 1
		}
		if err := q.UpsertTeacherNotificationPreference(ctx, queries.UpsertTeacherNotificationPreferenceParams{
			TeacherID: user.ID,
			Category:  string(def.Category),
			Enabled:   enabled,
		}); err != nil {
			logs.Log().Error("upsert notification preference", zap.Error(err), zap.String("category", string(def.Category)))
			setErrorFlash(w, "Failed to save notification preferences")
			HttpRedirect(w, r, "/profile?tab=notifications")
			return
		}
	}

	setSuccessFlash(w, "Notification preferences saved.")
	HttpRedirect(w, r, "/profile?tab=notifications")
}
