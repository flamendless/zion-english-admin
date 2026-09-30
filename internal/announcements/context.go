package announcements

import (
	"context"
	"database/sql"
	"net/http"

	"zion-english/internal/auth"
	"zion-english/internal/database/queries"
	"zion-english/internal/utils"
)

type contextKey string

const bannersKey contextKey = "announcement_banners"
const modalsKey contextKey = "announcement_modals"

type Banner struct {
	ID          int64
	Title       string
	Description string
	Level       string
	CTALabel    string
	CTAURL      string
}

type Modal struct {
	ID             int64
	Title          string
	Description    string
	Level          string
	CTALabel       string
	CTAURL         string
	ModalFrequency ModalFrequency
}

func GetBanners(ctx context.Context) []Banner {
	banners, _ := ctx.Value(bannersKey).([]Banner)
	return banners
}

func GetModals(ctx context.Context) []Modal {
	modals, _ := ctx.Value(modalsKey).([]Modal)
	return modals
}

func Middleware(db *queries.Queries, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := auth.GetUser(ctx)
		if user.Role == "" {
			next.ServeHTTP(w, r)
			return
		}

		today := utils.TodayPHT()
		var banners []Banner
		var modals []Modal

		if auth.HasAdminAccess(user.Role) {
			bannerRows, err := db.GetActiveAnnouncementsAll(ctx, today)
			if err == nil {
				banners = mapAllBannerRows(bannerRows, today)
			}
			modalRows, err := db.GetActiveModalAnnouncementsAll(ctx, today)
			if err == nil {
				modals = mapAllModalRows(modalRows, today)
			}
		} else if user.Role == auth.RoleTeacher && user.ID > 0 {
			bannerRows, err := db.GetActiveAnnouncementsForTeacher(ctx, queries.GetActiveAnnouncementsForTeacherParams{
				Date:      today,
				TeacherID: user.ID,
			})
			if err == nil {
				banners = mapTeacherBannerRows(bannerRows, today)
			}
			modalRows, err := db.GetActiveModalAnnouncementsForTeacher(ctx, queries.GetActiveModalAnnouncementsForTeacherParams{
				Date:      today,
				TeacherID: user.ID,
			})
			if err == nil {
				modals = mapTeacherModalRows(modalRows, today)
			}
		}

		ctx = context.WithValue(ctx, bannersKey, banners)
		ctx = context.WithValue(ctx, modalsKey, modals)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func mapAllBannerRows(rows []queries.GetActiveAnnouncementsAllRow, today string) []Banner {
	banners := make([]Banner, 0, len(rows))
	for _, row := range rows {
		if !isRowVisibleToday(row, today) {
			continue
		}
		banners = append(banners, mapBanner(row.ID, row.Title, row.Description, row.Level, row.CtaLabel, row.CtaUrl))
	}
	return banners
}

func mapTeacherBannerRows(rows []queries.GetActiveAnnouncementsForTeacherRow, today string) []Banner {
	banners := make([]Banner, 0, len(rows))
	for _, row := range rows {
		if !isRowVisibleToday(row, today) {
			continue
		}
		banners = append(banners, mapBanner(row.ID, row.Title, row.Description, row.Level, row.CtaLabel, row.CtaUrl))
	}
	return banners
}

func mapAllModalRows(rows []queries.GetActiveModalAnnouncementsAllRow, today string) []Modal {
	modals := make([]Modal, 0, len(rows))
	for _, row := range rows {
		if !isRowVisibleToday(row, today) {
			continue
		}
		modals = append(modals, mapModalRow(row))
	}
	return modals
}

func mapTeacherModalRows(rows []queries.GetActiveModalAnnouncementsForTeacherRow, today string) []Modal {
	modals := make([]Modal, 0, len(rows))
	for _, row := range rows {
		if !isRowVisibleToday(row, today) {
			continue
		}
		modals = append(modals, mapModalRow(row))
	}
	return modals
}

func isRowVisibleToday(row interface{}, today string) bool {
	switch r := row.(type) {
	case queries.GetActiveAnnouncementsAllRow:
		return IsVisibleOnDate(visibilityFromAllRow(r), today)
	case queries.GetActiveAnnouncementsForTeacherRow:
		return IsVisibleOnDate(visibilityFromTeacherRow(r), today)
	case queries.GetActiveModalAnnouncementsAllRow:
		return IsVisibleOnDate(visibilityFromModalAllRow(r), today)
	case queries.GetActiveModalAnnouncementsForTeacherRow:
		return IsVisibleOnDate(visibilityFromModalTeacherRow(r), today)
	default:
		return false
	}
}

func visibilityFromAllRow(row queries.GetActiveAnnouncementsAllRow) VisibilityInput {
	return VisibilityInputFromStorage(row.StartDate, row.EndDate, row.RepeatEnabled, row.RepeatSchedule, row.CutoffRepeatDays)
}

func visibilityFromTeacherRow(row queries.GetActiveAnnouncementsForTeacherRow) VisibilityInput {
	return VisibilityInputFromStorage(row.StartDate, row.EndDate, row.RepeatEnabled, row.RepeatSchedule, row.CutoffRepeatDays)
}

func visibilityFromModalAllRow(row queries.GetActiveModalAnnouncementsAllRow) VisibilityInput {
	return VisibilityInputFromStorage(row.StartDate, row.EndDate, row.RepeatEnabled, row.RepeatSchedule, row.CutoffRepeatDays)
}

func visibilityFromModalTeacherRow(row queries.GetActiveModalAnnouncementsForTeacherRow) VisibilityInput {
	return VisibilityInputFromStorage(row.StartDate, row.EndDate, row.RepeatEnabled, row.RepeatSchedule, row.CutoffRepeatDays)
}

func mapModalRow(row interface{}) Modal {
	switch r := row.(type) {
	case queries.GetActiveModalAnnouncementsAllRow:
		return mapModal(r.ID, r.Title, r.Description, r.Level, r.CtaLabel, r.CtaUrl, r.ModalFrequency)
	case queries.GetActiveModalAnnouncementsForTeacherRow:
		return mapModal(r.ID, r.Title, r.Description, r.Level, r.CtaLabel, r.CtaUrl, r.ModalFrequency)
	default:
		return Modal{}
	}
}

func mapModal(id int64, title, description, level, ctaLabel, ctaURL string, freq sql.NullString) Modal {
	frequency := ModalFrequencyPerSession
	if freq.Valid && ValidModalFrequency(freq.String) {
		frequency = ModalFrequency(freq.String)
	}
	return Modal{
		ID:             id,
		Title:          title,
		Description:    description,
		Level:          level,
		CTALabel:       ctaLabel,
		CTAURL:         ctaURL,
		ModalFrequency: frequency,
	}
}

func mapBanner(id int64, title, description, level, ctaLabel, ctaURL string) Banner {
	return Banner{
		ID:          id,
		Title:       title,
		Description: description,
		Level:       level,
		CTALabel:    ctaLabel,
		CTAURL:      ctaURL,
	}
}
