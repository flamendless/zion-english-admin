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

		if user.Role == auth.RoleSuperuser {
			rows, err := db.GetActiveAnnouncementsAll(ctx, today)
			if err == nil {
				banners = mapRows(rows)
			}
		} else if user.Role == auth.RoleTeacher && user.ID > 0 {
			rows, err := db.GetActiveAnnouncementsForTeacher(ctx, queries.GetActiveAnnouncementsForTeacherParams{
				Date:      today,
				TeacherID: user.ID,
			})
			if err == nil {
				banners = mapRows(rows)
			}
		}

		ctx = context.WithValue(ctx, bannersKey, banners)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func mapRows(rows []queries.TblAnnouncement) []Banner {
	banners := make([]Banner, 0, len(rows))
	for _, row := range rows {
		banners = append(banners, Banner{
			ID:          row.ID,
			Title:       row.Title,
			Description: row.Description,
			Level:       row.Level,
		})
	}
	return banners
}
