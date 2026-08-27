package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"zion-english/frontend"
	"zion-english/internal/announcements"
	"zion-english/internal/auth"
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

	today := utils.TodayPHT()
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
		isDeleted := row.Status == announcements.StatusDeleted
		items = append(items, frontend.AnnouncementListItem{
			ID:                strconv.FormatInt(row.ID, 10),
			Title:             row.Title,
			Level:             row.Level,
			StartDate:         row.StartDate,
			EndDate:           row.EndDate,
			Audience:          audience,
			Schedule:          announcementSchedule(today, row.StartDate, row.EndDate),
			ScheduleTone:      announcementScheduleTone(today, row.StartDate, row.EndDate),
			PublicationStatus: announcementPublicationStatus(row.Status),
			PublicationTone:   announcementPublicationTone(row.Status),
			IsDeleted:         isDeleted,
		})
	}

	filterPath := utils.URL("/announcements")
	data := frontend.AnnouncementListData{
		Announcements:  items,
		PageNumber:     page.Number,
		PageTotalPages: page.TotalPages(),
		PageTotal:      page.Total,
		PrevURL:        utils.BuildPageURLAt(filterPath, page.Number-1, page.Size, nil),
		NextURL:        utils.BuildPageURLAt(filterPath, page.Number+1, page.Size, nil),
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
			Today:     utils.TodayPHT(),
			VisibleTo: "all",
			Level:     announcements.LevelInfo,
			Status:    announcements.StatusDraft,
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
		setErrorFlash(w, "Invalid form data")
		HttpRedirect(w, r, "/announcements/register")
		return
	}

	req := parseAnnouncementRequest(r)
	if err := announcements.ValidateRequest(req, false); err != nil {
		setErrorFlash(w, err.Error())
		HttpRedirect(w, r, "/announcements/register")
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
		CtaLabel:     req.CTALabel,
		CtaUrl:       req.CTAURL,
		Status:       req.Status,
	})
	if err != nil {
		setErrorFlash(w, fmt.Sprintf("Failed to create announcement: %v", err))
		HttpRedirect(w, r, "/announcements/register")
		return
	}

	if !req.VisibleToAll {
		if err := announcements.ReplaceTeacherLinks(ctx, dbRW.GetQueries(), id, req.TeacherIDs); err != nil {
			setErrorFlash(w, fmt.Sprintf("Failed to save teacher visibility: %v", err))
			HttpRedirect(w, r, "/announcements/register")
			return
		}
	}

	insertAuditLogAs(ctx, auth.GetUser(ctx), "announcements", fmt.Sprintf("Created announcement #%d: %s", id, req.Title))
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
		if row.Status == announcements.StatusDeleted {
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

		formStatus := row.Status
		if !announcements.ValidFormStatus(formStatus) {
			formStatus = announcements.StatusDraft
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
			Today:       utils.TodayPHT(),
			IsEdit:      true,
			CTALabel:    row.CtaLabel,
			CTAURL:      row.CtaUrl,
			Status:      formStatus,
		})
	case http.MethodPost:
		handleAnnouncementUpdate(w, r, announcementID)
	default:
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleAnnouncementUpdate(w http.ResponseWriter, r *http.Request, announcementID int64) {
	ctx := r.Context()
	editPath := fmt.Sprintf("/announcements/%d/edit", announcementID)

	existing, err := dbRO.GetQueries().GetAnnouncementByID(ctx, announcementID)
	if err != nil {
		setErrorFlash(w, "Announcement not found")
		HttpRedirect(w, r, "/announcements")
		return
	}
	if existing.Status == announcements.StatusDeleted {
		setErrorFlash(w, "Announcement not found")
		HttpRedirect(w, r, "/announcements")
		return
	}

	if err := r.ParseForm(); err != nil {
		setErrorFlash(w, "Invalid form data")
		HttpRedirect(w, r, editPath)
		return
	}

	req := parseAnnouncementRequest(r)
	req.OriginalStart = existing.StartDate
	if err := announcements.ValidateRequest(req, true); err != nil {
		setErrorFlash(w, err.Error())
		HttpRedirect(w, r, editPath)
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
		CtaLabel:     req.CTALabel,
		CtaUrl:       req.CTAURL,
		Status:       req.Status,
		ID:           announcementID,
	}); err != nil {
		setErrorFlash(w, fmt.Sprintf("Failed to update announcement: %v", err))
		HttpRedirect(w, r, editPath)
		return
	}

	if req.VisibleToAll {
		if err := dbRW.GetQueries().DeleteAnnouncementTeacherLinks(ctx, announcementID); err != nil {
			setErrorFlash(w, fmt.Sprintf("Failed to update visibility: %v", err))
			HttpRedirect(w, r, editPath)
			return
		}
	} else if err := announcements.ReplaceTeacherLinks(ctx, dbRW.GetQueries(), announcementID, req.TeacherIDs); err != nil {
		setErrorFlash(w, fmt.Sprintf("Failed to save teacher visibility: %v", err))
		HttpRedirect(w, r, editPath)
		return
	}

	insertAuditLogAs(ctx, auth.GetUser(ctx), "announcements", fmt.Sprintf("Updated announcement #%d: %s", announcementID, req.Title))
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
	if row.Status == announcements.StatusDeleted {
		setErrorFlash(w, "Announcement is already deleted")
		HttpRedirect(w, r, "/announcements")
		return
	}

	if err := dbRW.GetQueries().DeleteAnnouncement(ctx, announcementID); err != nil {
		setErrorFlash(w, fmt.Sprintf("Failed to delete announcement: %v", err))
		HttpRedirect(w, r, "/announcements")
		return
	}

	insertAuditLogAs(ctx, auth.GetUser(ctx), "announcements", fmt.Sprintf("Deleted announcement #%d: %s", announcementID, row.Title))
	setSuccessFlash(w, "Announcement deleted successfully")
	HttpRedirect(w, r, "/announcements")
}

func parseAnnouncementRequest(r *http.Request) announcements.Request {
	visibleToAll := r.FormValue("visible_to") != "selected"
	var teacherIDs []int64
	if !visibleToAll {
		teacherIDs, _ = parseAssignedTeacherIDs(r.Context(), dbRO.GetQueries(), r.Form["teachers"])
	}
	status := r.FormValue("status")
	if status == "" {
		status = announcements.StatusDraft
	}
	return announcements.Request{
		Title:        r.FormValue("title"),
		Description:  r.FormValue("description"),
		Level:        r.FormValue("level"),
		StartDate:    r.FormValue("start_date"),
		EndDate:      r.FormValue("end_date"),
		VisibleToAll: visibleToAll,
		TeacherIDs:   teacherIDs,
		CTALabel:     r.FormValue("cta_label"),
		CTAURL:       r.FormValue("cta_url"),
		Status:       status,
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
			Name:     utils.ComposePersonName(t.FirstName, t.MiddleName, t.LastName),
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

func announcementSchedule(today, start, end string) string {
	if today < start {
		return "Upcoming"
	}
	if today > end {
		return "Expired"
	}
	return "Active"
}

func announcementScheduleTone(today, start, end string) frontend.PillTone {
	return frontend.AnnouncementSchedulePillTone(announcementSchedule(today, start, end))
}

func announcementPublicationStatus(status string) string {
	switch status {
	case announcements.StatusPublished:
		return "Published"
	case announcements.StatusDeleted:
		return "Deleted"
	default:
		return "Draft"
	}
}

func announcementPublicationTone(status string) frontend.PillTone {
	return frontend.AnnouncementPublicationPillTone(status)
}
