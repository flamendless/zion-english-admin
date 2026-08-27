package announcements

import (
	"context"
	"net/http"

	"zion-english/internal/auth"
	"zion-english/internal/database/queries"
	"zion-english/internal/utils"
)

type contextKey string

const bannersKey contextKey = "announcement_banners"

type Banner struct {
	ID          int64
	Title       string
	Description string
	Level       string
	CTALabel    string
	CTAURL      string
}

func GetBanners(ctx context.Context) []Banner {
	banners, _ := ctx.Value(bannersKey).([]Banner)
	return banners
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

		if auth.HasAdminAccess(user.Role) {
			rows, err := db.GetActiveAnnouncementsAll(ctx, today)
			if err == nil {
				banners = mapAllRows(rows)
			}
		} else if user.Role == auth.RoleTeacher && user.ID > 0 {
			rows, err := db.GetActiveAnnouncementsForTeacher(ctx, queries.GetActiveAnnouncementsForTeacherParams{
				Date:      today,
				TeacherID: user.ID,
			})
			if err == nil {
				banners = mapTeacherRows(rows)
			}
		}

		ctx = context.WithValue(ctx, bannersKey, banners)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func mapAllRows(rows []queries.GetActiveAnnouncementsAllRow) []Banner {
	banners := make([]Banner, 0, len(rows))
	for _, row := range rows {
		banners = append(banners, mapBanner(row.ID, row.Title, row.Description, row.Level, row.CtaLabel, row.CtaUrl))
	}
	return banners
}

func mapTeacherRows(rows []queries.GetActiveAnnouncementsForTeacherRow) []Banner {
	banners := make([]Banner, 0, len(rows))
	for _, row := range rows {
		banners = append(banners, mapBanner(row.ID, row.Title, row.Description, row.Level, row.CtaLabel, row.CtaUrl))
	}
	return banners
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
