package cmd

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/conf"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/utils"

	"go.uber.org/zap"
)

func weekRange() (string, string) {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	start := now.AddDate(0, 0, -(weekday - 1))
	end := start.AddDate(0, 0, 6)
	return start.Format("2006-01-02"), end.Format("2006-01-02")
}

func monthRange() (string, string) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 1, -1)
	return start.Format("2006-01-02"), end.Format("2006-01-02")
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	role := auth.GetRole(r.Context())
	data := frontend.DashboardData{Role: role}

	weekStart, weekEnd := weekRange()
	monthStart, monthEnd := monthRange()

	if role == auth.RoleSuperuser {
		statusCounts, err := dbRO.GetQueries().CountStudentsByStatus(ctx)
		if err == nil {
			for _, row := range statusCounts {
				switch row.Status {
				case "active":
					data.ActiveStudents = row.Count
				case "inactive":
					data.InactiveStudents = row.Count
				}
			}
		}
		pending, err := dbRO.GetQueries().CountTeachersByStatus(ctx, "pending")
		if err == nil {
			data.PendingTeachers = pending
		}
		classCounts, err := dbRO.GetQueries().CountClassRecordsByStatusAndDateRange(ctx, queries.CountClassRecordsByStatusAndDateRangeParams{
			Date:      weekStart,
			Date_2:    weekEnd,
			Column3:   int64(0),
			TeacherID: 0,
		})
		if err == nil {
			for _, row := range classCounts {
				switch row.Status {
				case "conducted":
					data.ClassesConductedWeek = row.Count
				case "cancelled":
					data.ClassesCancelledWeek = row.Count
				case "rescheduled":
					data.ClassesRescheduledWeek = row.Count
				}
			}
		}
		rates, err := dbRO.GetQueries().SumConductedRateByCurrencyAndDateRange(ctx, queries.SumConductedRateByCurrencyAndDateRangeParams{
			Date:      monthStart,
			Date_2:    monthEnd,
			Column3:   int64(0),
			TeacherID: 0,
		})
		if err == nil {
			for _, row := range rates {
				total, _ := row.TotalRate.(float64)
				data.MonthlyTotals = append(data.MonthlyTotals, frontend.CurrencyTotal{
					Currency: row.Currency,
					Total:    total,
				})
			}
		}
	} else if role == auth.RoleTeacher {
		user := auth.GetUser(ctx)
		count, err := dbRO.GetQueries().CountStudentsByTeacherID(ctx, user.ID)
		if err == nil {
			data.MyStudentCount = count
		}
		classCounts, err := dbRO.GetQueries().CountClassRecordsByStatusAndDateRange(ctx, queries.CountClassRecordsByStatusAndDateRangeParams{
			Date:      weekStart,
			Date_2:    weekEnd,
			Column3:   user.ID,
			TeacherID: user.ID,
		})
		if err == nil {
			for _, row := range classCounts {
				switch row.Status {
				case "conducted":
					data.ClassesConductedWeek = row.Count
				case "cancelled":
					data.ClassesCancelledWeek = row.Count
				case "rescheduled":
					data.ClassesRescheduledWeek = row.Count
				}
			}
		}
		rates, err := dbRO.GetQueries().SumConductedRateByCurrencyAndDateRange(ctx, queries.SumConductedRateByCurrencyAndDateRangeParams{
			Date:      monthStart,
			Date_2:    monthEnd,
			Column3:   user.ID,
			TeacherID: user.ID,
		})
		if err == nil {
			for _, row := range rates {
				total, _ := row.TotalRate.(float64)
				data.MonthlyTotals = append(data.MonthlyTotals, frontend.CurrencyTotal{
					Currency: row.Currency,
					Total:    total,
				})
			}
		}
		today := time.Now().Format("2006-01-02")
		scheduledToday, err := dbRO.GetQueries().CountScheduledClassesByStatusAndDate(ctx, queries.CountScheduledClassesByStatusAndDateParams{
			ScheduledDate: today,
			Status:        "scheduled",
			Column3:       user.ID,
			TeacherID:     user.ID,
		})
		if err == nil {
			data.ScheduledToday = scheduledToday
		}
	}

	if err := frontend.Home(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleMyStudents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if user.ID == 0 {
		HttpError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	q := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")
	page := utils.ParsePageQuery(r)

	filter := queries.CountStudentsByTeacherIDFilteredParams{
		TeacherID: user.ID,
		Column2:   q,
		Column3:   sql.NullString{String: q, Valid: true},
		Column4:   status,
		Status:    status,
	}

	total, err := dbRO.GetQueries().CountStudentsByTeacherIDFiltered(ctx, filter)
	if err != nil {
		HttpError(w, "Failed to count students", http.StatusInternalServerError)
		return
	}
	page.Total = total

	students, err := dbRO.GetQueries().GetStudentsByTeacherIDFiltered(ctx, queries.GetStudentsByTeacherIDFilteredParams{
		TeacherID: user.ID,
		Column2:   q,
		Column3:   sql.NullString{String: q, Valid: true},
		Column4:   status,
		Status:    status,
		Limit:     int64(page.Size),
		Offset:    int64(page.Offset()),
	})
	if err != nil {
		HttpError(w, "Failed to fetch students", http.StatusInternalServerError)
		return
	}

	viewStudents := make([]frontend.StudentItem, len(students))
	for i, s := range students {
		viewStudents[i] = frontend.StudentItem{
			ID:            strconv.FormatInt(s.ID, 10),
			Name:          s.Name,
			Currency:      s.Currency,
			Contact:       s.Contact.String,
			RatePerClass:  s.RatePerClass,
			ParentName:    s.ParentName.String,
			AssignedColor: s.AssignedColor,
			Status:        s.Status,
		}
	}

	params := map[string]string{"q": q, "status": status}
	w.Header().Set("Content-Type", "text/html")
	frontend.MyStudents(frontend.MyStudentsData{
		Students:       viewStudents,
		Query:          q,
		Status:         status,
		PageNumber:     page.Number,
		PageTotalPages: page.TotalPages(),
		PageTotal:      page.Total,
		PrevURL:        utils.BuildPageURLAt(utils.URL("/my-students"), page.Number-1, page.Size, params),
		NextURL:        utils.BuildPageURLAt(utils.URL("/my-students"), page.Number+1, page.Size, params),
		HasPrev:        page.HasPrev(),
		HasNext:        page.HasNext(),
		FilterPath:     utils.URL("/my-students"),
	}).Render(ctx, w)
}

func handleLogoutWithAccess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	if user, ok := auth.UserFromRequest(r, conf.Conf()); ok {
		if user.Role == auth.RoleSuperuser {
			insertAuditLogAs(ctx, user, "auth", fmt.Sprintf("logged out (%s)", user.Email))
		} else if user.Role == auth.RoleTeacher && user.ID > 0 {
			insertAuditLogAs(ctx, user, "auth", fmt.Sprintf("logged out (%s)", user.Email))
			accessID, err := dbRW.GetQueries().GetLatestOpenAccessByTeacherID(ctx, user.ID)
			if err == nil && accessID > 0 {
				if _, err := dbRW.GetQueries().UpdateAccessLogout(ctx, accessID); err != nil {
					logs.Log().Warn("[Handle Logout]", zap.Error(err))
				}
			}
		}
	}

	auth.Logout(w)
	HttpRedirect(w, r, "/")
}
