package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/conf"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/onboarding"
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

func dashboardEarningsTeacherID(role auth.Role, userID int64) int64 {
	if role == auth.RoleSuperuser {
		return 0
	}
	return userID
}

func fetchCutoffTotalsByPreset(ctx context.Context, preset string, teacherID int64) []frontend.CurrencyTotal {
	startDate, endDate := utils.CutoffDatesFromPreset(preset)
	if startDate == "" || endDate == "" {
		return nil
	}
	rows, err := dbRO.GetQueries().SumConductedRateByCurrencyAndDateRange(ctx, queries.SumConductedRateByCurrencyAndDateRangeParams{
		Date:      startDate,
		Date_2:    endDate,
		Column3:   teacherID,
		TeacherID: teacherID,
	})
	if err != nil {
		return nil
	}
	totals := make([]frontend.CurrencyTotal, 0, len(rows))
	for _, row := range rows {
		total, _ := row.TotalRate.(float64)
		totals = append(totals, frontend.CurrencyTotal{
			Currency: row.Currency,
			Total:    total,
		})
	}
	return totals
}

func fetchCutoffParentTotalsByPreset(ctx context.Context, preset string, teacherID int64) []frontend.CurrencyTotal {
	startDate, endDate := utils.CutoffDatesFromPreset(preset)
	if startDate == "" || endDate == "" {
		return nil
	}
	rows, err := dbRO.GetQueries().SumConductedParentRateByCurrencyAndDateRange(ctx, queries.SumConductedParentRateByCurrencyAndDateRangeParams{
		Date:      startDate,
		Date_2:    endDate,
		Column3:   teacherID,
		TeacherID: teacherID,
	})
	if err != nil {
		return nil
	}
	totals := make([]frontend.CurrencyTotal, 0, len(rows))
	for _, row := range rows {
		if !row.Currency.Valid || row.Currency.String == "" {
			continue
		}
		total, _ := row.TotalRate.(float64)
		totals = append(totals, frontend.CurrencyTotal{
			Currency: row.Currency.String,
			Total:    total,
		})
	}
	return totals
}

func populateAllTeachersCutoffTotals(ctx context.Context, data *frontend.DashboardData) {
	if data.Role != auth.RoleAdmin {
		return
	}
	firstCutoff, secondCutoff, _ := utils.CurrentCutoffRange()
	data.AllTeachersFirstCutoffTotals = fetchCutoffTotalsByPreset(ctx, firstCutoff, 0)
	data.AllTeachersSecondCutoffTotals = fetchCutoffTotalsByPreset(ctx, secondCutoff, 0)
	data.AllTeachersFirstCutoffParentTotals = fetchCutoffParentTotalsByPreset(ctx, firstCutoff, 0)
	data.AllTeachersSecondCutoffParentTotals = fetchCutoffParentTotalsByPreset(ctx, secondCutoff, 0)
}

func populateDashboardEarnings(ctx context.Context, data *frontend.DashboardData, teacherID int64) {
	firstCutoff, secondCutoff, _ := utils.CurrentCutoffRange()
	data.FirstCutoffTotals = fetchCutoffTotalsByPreset(ctx, firstCutoff, teacherID)
	data.SecondCutoffTotals = fetchCutoffTotalsByPreset(ctx, secondCutoff, teacherID)
	if !auth.HasAdminAccess(data.Role) {
		return
	}
	data.FirstCutoffParentTotals = fetchCutoffParentTotalsByPreset(ctx, firstCutoff, teacherID)
	data.SecondCutoffParentTotals = fetchCutoffParentTotalsByPreset(ctx, secondCutoff, teacherID)
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	role := auth.GetRole(r.Context())
	user := auth.GetUser(ctx)
	data := frontend.DashboardData{Role: role}

	avatar, err := buildHeaderAvatarProps(ctx, user, role)
	if err != nil {
		logs.Log().Error("build dashboard avatar", zap.Error(err))
	} else {
		data.Avatar = avatar
	}

	weekStart, weekEnd := weekRange()
	earningsTeacherID := dashboardEarningsTeacherID(role, user.ID)

	switch role {
	case auth.RoleSuperuser, auth.RoleAdmin:
		statusCounts, err := dbRO.GetQueries().CountStudentsByStatus(ctx)
		if err == nil {
			for _, row := range statusCounts {
				switch constants.StudentStatus(row.Status) {
				case constants.StudentStatusActive:
					data.ActiveStudents = row.Count
				case constants.StudentStatusInactive:
					data.InactiveStudents = row.Count
				}
			}
		}
		pending, err := dbRO.GetQueries().CountTeachersByStatus(ctx, "pending")
		if err == nil {
			data.PendingTeachers = pending
		}
		pendingDocs, err := dbRO.GetQueries().CountTeacherDocumentsByStatus(ctx, string(constants.TeacherDocumentStatusSubmitted))
		if err == nil {
			data.PendingDocuments = pendingDocs
		}
		pendingIntroVideos, err := dbRO.GetQueries().CountTeacherIntroVideosByStatus(ctx, string(constants.TeacherIntroVideoStatusSubmitted))
		if err == nil {
			data.PendingIntroVideos = pendingIntroVideos
		}
		withoutParent, err := dbRO.GetQueries().CountActiveStudentsWithoutParent(ctx)
		if err == nil {
			data.StudentsWithoutParent = withoutParent
		}
		withoutParentRate, err := dbRO.GetQueries().CountActiveStudentsWithoutParentRate(ctx)
		if err == nil {
			data.StudentsWithoutParentRate = withoutParentRate
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
		populateDashboardEarnings(ctx, &data, earningsTeacherID)
		populateAllTeachersCutoffTotals(ctx, &data)
	case auth.RoleTeacher, auth.RoleTester:
		user := auth.GetUser(ctx)
		if role == auth.RoleTeacher {
			checklist, err := loadTeacherOnboardingChecklist(ctx, user.ID)
			if err == nil {
				completed, total := onboarding.Summary(checklist)
				data.ShowOnboarding = total > 0 && completed < total
				data.Onboarding = frontend.OnboardingChecklistData{
					Items:          frontend.MapOnboardingItems(checklist),
					CompletedCount: completed,
					TotalCount:     total,
					ShowSummary:    true,
				}
			}
			populatePaymentReceipt(ctx, &data, user.ID)
		}
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
		populateDashboardEarnings(ctx, &data, earningsTeacherID)
		today := utils.TodayPHT()
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
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if user.ID == 0 {
		HttpError(w, MsgUnauthorized, http.StatusUnauthorized)
		return
	}

	q := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")
	sort := parseListSort(r, frontend.ListSortKindMyStudent)
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

	allStudents, err := dbRO.GetQueries().GetStudentsByTeacherIDFiltered(ctx, queries.GetStudentsByTeacherIDFilteredParams{
		TeacherID: user.ID,
		Column2:   q,
		Column3:   sql.NullString{String: q, Valid: true},
		Column4:   status,
		Status:    status,
		Limit:     total,
		Offset:    0,
	})
	if err != nil {
		HttpError(w, MsgFailedToFetchStudents, http.StatusInternalServerError)
		return
	}
	sortMyStudentRows(allStudents, sort)
	students := paginateSlice(allStudents, page)

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
			Status:        constants.StudentStatus(s.Status),
		}
	}

	params := listQueryParamsWithSort(r, frontend.ListSortKindMyStudent)
	writeHTML(w)
	frontend.MyStudents(frontend.MyStudentsData{
		Students:       viewStudents,
		Query:          q,
		Status:         constants.StudentStatus(status),
		SortBy:         sort.By,
		SortOrder:      string(sort.Order),
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
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	if user, ok := auth.UserFromRequest(r, conf.Conf()); ok {
		if auth.HasAdminAccess(user.Role) {
			insertAuditLogAs(ctx, user, "auth", fmt.Sprintf("logged out (%s)", user.Email))
		} else if auth.IsTeacherScoped(user.Role) && user.ID > 0 {
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
	HttpRedirect(w, r, "/auth/login")
}
