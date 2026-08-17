package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"zion-english/frontend"
	"zion-english/internal/announcements"
	"zion-english/internal/database/queries"
	"zion-english/internal/utils"
)

func handleAnnouncementsPath(w http.ResponseWriter, r *http.Request) {
	if id, ok := extractPathID(r, "announcements", "/edit"); ok {
		handleAnnouncementEdit(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "announcements", "/delete"); ok {
		handleAnnouncementDelete(w, r, id)
		return
	}
	HttpError(w, "Not found", http.StatusNotFound)
}

func handleAnnouncements(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	page := utils.ParsePageQuery(r)
	total, err := dbRO.GetQueries().CountAnnouncements(ctx)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to count announcements: %v", err), http.StatusInternalServerError)
		return
	}
	page.Total = total

	rows, err := dbRO.GetQueries().GetAnnouncementsPaged(ctx, queries.GetAnnouncementsPagedParams{
		Limit:  int64(page.Size),
		Offset: int64(page.Offset()),
	})
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to load announcements: %v", err), http.StatusInternalServerError)
		return
	}

	today := todayPHTString()
	items := make([]frontend.AnnouncementListItem, 0, len(rows))
	for _, row := range rows {
		audience := "All teachers"
		if row.VisibleToAll == 0 {
			count, err := dbRO.GetQueries().CountTeachersByAnnouncementID(ctx, row.ID)
			if err == nil {
				audience = fmt.Sprintf("%d teacher(s)", count)
			} else {
				audience = "Selected teachers"
			}
		}
		items = append(items, frontend.AnnouncementListItem{
			ID:          strconv.FormatInt(row.ID, 10),
			Title:       row.Title,
			Level:       row.Level,
			StartDate:   row.StartDate,
			EndDate:     row.EndDate,
			Audience:    audience,
			Status:      announcementStatus(today, row.StartDate, row.EndDate),
			StatusClass: announcementStatusClass(today, row.StartDate, row.EndDate),
		})
	}

	filterPath := utils.URL("/announcements")
	data := frontend.AnnouncementListData{
		Announcements:  items,
		PageNumber:     page.Number,
		PageTotalPages: page.TotalPages(),
		PageTotal:      page.Total,
		PrevURL:        pageURL(filterPath, page.Number-1, page.Size, nil),
		NextURL:        pageURL(filterPath, page.Number+1, page.Size, nil),
		HasPrev:        page.HasPrev(),
		HasNext:        page.HasNext(),
	}

	if err := frontend.Announcements(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleAnnouncementRegister(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		teachers, err := loadAnnouncementTeacherOptions(r.Context(), nil)
		if err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		renderAnnouncementForm(w, r, frontend.AnnouncementFormData{
			Today:     todayPHTString(),
			VisibleTo: "all",
			Level:     announcements.LevelInfo,
			Teachers:  teachers,
		})
	case http.MethodPost:
		handleAnnouncementCreate(w, r)
	default:
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleAnnouncementCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		HttpError(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	req := parseAnnouncementRequest(r)
	if err := announcements.ValidateRequest(req, false); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	visibleToAll := int64(1)
	if !req.VisibleToAll {
		visibleToAll = 0
	}

	id, err := dbRW.GetQueries().InsertAnnouncement(ctx, queries.InsertAnnouncementParams{
		Title:        req.Title,
		Description:  req.Description,
		Level:        req.Level,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		VisibleToAll: visibleToAll,
	})
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to create announcement: %v", err), http.StatusInternalServerError)
		return
	}

	if !req.VisibleToAll {
		if err := announcements.ReplaceTeacherLinks(ctx, dbRW.GetQueries(), id, req.TeacherIDs); err != nil {
			HttpError(w, fmt.Sprintf("Failed to save teacher visibility: %v", err), http.StatusInternalServerError)
			return
		}
	}

	insertAuditLog(ctx, "announcements", fmt.Sprintf("Created announcement #%d: %s", id, req.Title))
	setSuccessFlash(w, "Announcement created successfully")
	HttpRedirect(w, r, "/announcements")
}

func handleAnnouncementEdit(w http.ResponseWriter, r *http.Request, announcementID int64) {
	switch r.Method {
	case http.MethodGet:
		ctx := r.Context()
		row, err := dbRO.GetQueries().GetAnnouncementByID(ctx, announcementID)
		if err != nil {
			HttpError(w, "Announcement not found", http.StatusNotFound)
			return
		}

		teacherIDs, err := dbRO.GetQueries().GetTeacherIDsByAnnouncementID(ctx, announcementID)
		if err != nil {
			HttpError(w, fmt.Sprintf("Failed to load teachers: %v", err), http.StatusInternalServerError)
			return
		}

		selected := make([]string, 0, len(teacherIDs))
		for _, tid := range teacherIDs {
			selected = append(selected, strconv.FormatInt(tid, 10))
		}

		visibleTo := "all"
		if row.VisibleToAll == 0 {
			visibleTo = "selected"
		}

		teachers, err := loadAnnouncementTeacherOptions(ctx, selected)
		if err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		renderAnnouncementForm(w, r, frontend.AnnouncementFormData{
			ID:          strconv.FormatInt(announcementID, 10),
			Title:       row.Title,
			Description: row.Description,
			Level:       row.Level,
			StartDate:   row.StartDate,
			EndDate:     row.EndDate,
			VisibleTo:   visibleTo,
			Teachers:    teachers,
			Today:       todayPHTString(),
			IsEdit:      true,
		})
	case http.MethodPost:
		handleAnnouncementUpdate(w, r, announcementID)
	default:
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleAnnouncementUpdate(w http.ResponseWriter, r *http.Request, announcementID int64) {
	ctx := r.Context()
	existing, err := dbRO.GetQueries().GetAnnouncementByID(ctx, announcementID)
	if err != nil {
		HttpError(w, "Announcement not found", http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		HttpError(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	req := parseAnnouncementRequest(r)
	req.OriginalStart = existing.StartDate
	if err := announcements.ValidateRequest(req, true); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	visibleToAll := int64(1)
	if !req.VisibleToAll {
		visibleToAll = 0
	}

	if err := dbRW.GetQueries().UpdateAnnouncement(ctx, queries.UpdateAnnouncementParams{
		Title:        req.Title,
		Description:  req.Description,
		Level:        req.Level,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		VisibleToAll: visibleToAll,
		ID:           announcementID,
	}); err != nil {
		HttpError(w, fmt.Sprintf("Failed to update announcement: %v", err), http.StatusInternalServerError)
		return
	}

	if req.VisibleToAll {
		if err := dbRW.GetQueries().DeleteAnnouncementTeacherLinks(ctx, announcementID); err != nil {
			HttpError(w, fmt.Sprintf("Failed to update visibility: %v", err), http.StatusInternalServerError)
			return
		}
	} else if err := announcements.ReplaceTeacherLinks(ctx, dbRW.GetQueries(), announcementID, req.TeacherIDs); err != nil {
		HttpError(w, fmt.Sprintf("Failed to save teacher visibility: %v", err), http.StatusInternalServerError)
		return
	}

	insertAuditLog(ctx, "announcements", fmt.Sprintf("Updated announcement #%d: %s", announcementID, req.Title))
	setSuccessFlash(w, "Announcement updated successfully")
	HttpRedirect(w, r, "/announcements")
}

func handleAnnouncementDelete(w http.ResponseWriter, r *http.Request, announcementID int64) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	row, err := dbRO.GetQueries().GetAnnouncementByID(ctx, announcementID)
	if err != nil {
		HttpError(w, "Announcement not found", http.StatusNotFound)
		return
	}

	if err := dbRW.GetQueries().DeleteAnnouncement(ctx, announcementID); err != nil {
		HttpError(w, fmt.Sprintf("Failed to delete announcement: %v", err), http.StatusInternalServerError)
		return
	}

	insertAuditLog(ctx, "announcements", fmt.Sprintf("Deleted announcement #%d: %s", announcementID, row.Title))
	setSuccessFlash(w, "Announcement deleted successfully")
	HttpRedirect(w, r, "/announcements")
}

func parseAnnouncementRequest(r *http.Request) announcements.Request {
	visibleToAll := r.FormValue("visible_to") != "selected"
	var teacherIDs []int64
	if !visibleToAll {
		teacherIDs, _ = parseAssignedTeacherIDs(r.Context(), dbRO.GetQueries(), r.Form["teachers"])
	}
	return announcements.Request{
		Title:        strings.TrimSpace(r.FormValue("title")),
		Description:  strings.TrimSpace(r.FormValue("description")),
		Level:        r.FormValue("level"),
		StartDate:    r.FormValue("start_date"),
		EndDate:      r.FormValue("end_date"),
		VisibleToAll: visibleToAll,
		TeacherIDs:   teacherIDs,
	}
}

func loadAnnouncementTeacherOptions(ctx context.Context, selected []string) ([]frontend.AnnouncementTeacherOption, error) {
	rows, err := dbRO.GetQueries().GetApprovedTeachers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load teachers: %w", err)
	}
	selectedSet := make(map[string]bool, len(selected))
	for _, id := range selected {
		selectedSet[id] = true
	}
	options := make([]frontend.AnnouncementTeacherOption, 0, len(rows))
	for _, t := range rows {
		id := strconv.FormatInt(t.ID, 10)
		options = append(options, frontend.AnnouncementTeacherOption{
			ID:       id,
			Name:     t.Name,
			Selected: selectedSet[id],
		})
	}
	return options, nil
}

func renderAnnouncementForm(w http.ResponseWriter, r *http.Request, data frontend.AnnouncementFormData) {
	ctx := r.Context()
	if err := frontend.AnnouncementForm(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func todayPHTString() string {
	return utils.DatePHT(time.Now())
}

func announcementStatus(today, start, end string) string {
	if today < start {
		return "Upcoming"
	}
	if today > end {
		return "Expired"
	}
	return "Active"
}

func announcementStatusClass(today, start, end string) string {
	switch announcementStatus(today, start, end) {
	case "Active":
		return "status-active"
	case "Upcoming":
		return "status-pending"
	default:
		return "status-inactive"
	}
}
