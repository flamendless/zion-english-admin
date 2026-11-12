package payroll

import (
	"context"
	"net/http"

	"zion-english/internal/auth"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
)

type contextKey string

const bannerKey contextKey = "payroll_session_banner"

type Banner struct {
	SessionID              int64
	PeriodLabel            string
	NeedsDailyLogsConfirm  bool
}

func GetBanner(ctx context.Context) *Banner {
	b, _ := ctx.Value(bannerKey).(*Banner)
	return b
}

func Middleware(db *queries.Queries, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := auth.GetUser(ctx)
		if user.Role == "" || auth.HasAdminAccess(user.Role) {
			next.ServeHTTP(w, r)
			return
		}
		row, err := db.GetActivePayrollSessionTeacherRow(ctx, user.ID)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		ctx = context.WithValue(ctx, bannerKey, &Banner{
			SessionID:             row.SessionID,
			PeriodLabel:           CutoffLabel(row.PeriodStart, row.PeriodEnd),
			NeedsDailyLogsConfirm: row.Status == string(constants.PayrollTeacherStatusAccomplishDailyLogs),
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
