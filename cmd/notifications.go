package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/utils"
)

func handleNotificationsPath(w http.ResponseWriter, r *http.Request) {
	if id, ok := extractPathID(r, "notifications", "/read"); ok {
		handleNotificationRead(w, r, id)
		return
	}
	HttpError(w, "Not found", http.StatusNotFound)
}

func handleNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	notifySvc.ScanMissedClasses(ctx)

	unreadOnly := r.URL.Query().Get("filter") == "unread"
	sort := parseListSort(r, frontend.ListSortKindNotification)
	page := utils.ParsePageQuery(r)
	total, err := notifySvc.Count(ctx, user, unreadOnly)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to count notifications: %v", err), http.StatusInternalServerError)
		return
	}
	page.Total = total

	allRows, err := notifySvc.ListPaged(ctx, user, unreadOnly, total, 0)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to load notifications: %v", err), http.StatusInternalServerError)
		return
	}
	sortNotificationRows(allRows, sort)
	rows := paginateSlice(allRows, page)

	params := map[string]string{"filter": r.URL.Query().Get("filter")}
	for k, v := range sort.QueryValues() {
		if v != "" {
			params[k] = v
		}
	}
	filterPath := utils.URL("/notifications")
	data := frontend.NotificationListData{
		Items:          notificationItems(rows),
		UnreadOnly:     unreadOnly,
		SortBy:         sort.By,
		SortOrder:      string(sort.Order),
		PageNumber:     page.Number,
		PageTotalPages: page.TotalPages(),
		PageTotal:      page.Total,
		PrevURL:        utils.BuildPageURLAt(filterPath, page.Number-1, page.Size, params),
		NextURL:        utils.BuildPageURLAt(filterPath, page.Number+1, page.Size, params),
		HasPrev:        page.HasPrev(),
		HasNext:        page.HasNext(),
		FilterPath:     filterPath,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := frontend.Notifications(data).Render(ctx, w); err != nil {
		HttpError(w, fmt.Sprintf("Failed to render notifications: %v", err), http.StatusInternalServerError)
	}
}

func handleNotificationsPanel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	renderNotificationsPanel(w, r)
}

func handleNotificationsUnreadCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	count, err := notifySvc.UnreadCount(ctx, user)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to count notifications: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := frontend.NotificationBadge(count).Render(ctx, w); err != nil {
		HttpError(w, fmt.Sprintf("Failed to render badge: %v", err), http.StatusInternalServerError)
	}
}

func handleNotificationRead(w http.ResponseWriter, r *http.Request, id int64) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if err := notifySvc.MarkRead(ctx, user, id); err != nil {
		HttpError(w, fmt.Sprintf("Failed to mark notification read: %v", err), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", `{"notificationUpdated":"body"}`)
		if r.URL.Query().Get("panel") == "1" {
			renderNotificationsPanel(w, r)
			return
		}
		row, err := getNotificationForUser(ctx, user, id)
		if err != nil {
			HttpError(w, "Notification not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		item := notificationItems([]queries.TblNotification{row})[0]
		if err := frontend.NotificationListRow(item).Render(ctx, w); err != nil {
			HttpError(w, fmt.Sprintf("Failed to render notification: %v", err), http.StatusInternalServerError)
		}
		return
	}
	HttpRedirect(w, r, "/notifications")
}

func handleNotificationsReadAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if err := notifySvc.MarkAllRead(ctx, user); err != nil {
		HttpError(w, fmt.Sprintf("Failed to mark notifications read: %v", err), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", `{"notificationUpdated":"body"}`)
		if r.URL.Query().Get("panel") == "1" {
			renderNotificationsPanel(w, r)
			return
		}
		w.Header().Set("HX-Redirect", utils.URL("/notifications"))
		return
	}
	HttpRedirect(w, r, "/notifications")
}

func renderNotificationsPanel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := auth.GetUser(ctx)
	notifySvc.ScanMissedClasses(ctx)

	rows, err := notifySvc.Recent(ctx, user)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to load notifications: %v", err), http.StatusInternalServerError)
		return
	}
	unread, err := notifySvc.UnreadCount(ctx, user)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to count notifications: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := frontend.NotificationPanel(frontend.NotificationPanelData{
		Items:       notificationItems(rows),
		UnreadCount: unread,
	}).Render(ctx, w); err != nil {
		HttpError(w, fmt.Sprintf("Failed to render notification panel: %v", err), http.StatusInternalServerError)
	}
}

func getNotificationForUser(ctx context.Context, user auth.User, id int64) (queries.TblNotification, error) {
	if auth.HasAdminAccess(user.Role) {
		return dbRO.GetQueries().GetNotificationForSuperuser(ctx, id)
	}
	return dbRO.GetQueries().GetNotificationForTeacher(ctx, queries.GetNotificationForTeacherParams{
		ID:          id,
		ToTeacherID: user.ID,
	})
}

func notificationItems(rows []queries.TblNotification) []frontend.NotificationItem {
	items := make([]frontend.NotificationItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, frontend.NotificationItem{
			ID:        strconv.FormatInt(row.ID, 10),
			From:      row.FromName,
			To:        row.ToName,
			Message:   row.Message,
			CreatedAt: formatNotificationCreatedAt(row.CreatedAt),
			Read:      row.Read == 1,
		})
	}
	return items
}

func formatNotificationCreatedAt(value string) string {
	if value == "" {
		return "-"
	}
	layouts := []string{constants.DateTimeSecondsLayout, constants.DateTimeLayout}
	for _, layout := range layouts {
		t, err := time.Parse(layout, value)
		if err == nil {
			return utils.DateTimeSecondsPHT(t.UTC())
		}
	}
	return value
}
