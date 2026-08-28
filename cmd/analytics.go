package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/models"
	"zion-english/internal/utils"

	"go.uber.org/zap"
)

type analyticsSummaryJSON struct {
	Conducted         int64   `json:"conducted"`
	Cancelled         int64   `json:"cancelled"`
	Rescheduled       int64   `json:"rescheduled"`
	CancellationRate  float64 `json:"cancellationRate"`
	ScheduledMinutes  int64   `json:"scheduledMinutes"`
	ConductedMinutes  int64   `json:"conductedMinutes"`
	UtilizationPct    float64 `json:"utilizationPct"`
	NoShowCount       int64   `json:"noShowCount"`
}

type analyticsTeacherRowJSON struct {
	TeacherID        string            `json:"teacherId"`
	TeacherName      string            `json:"teacherName"`
	TeacherAvatar    models.AvatarView `json:"teacherAvatar"`
	Conducted        int64             `json:"conducted"`
	Cancelled        int64   `json:"cancelled"`
	Rescheduled      int64   `json:"rescheduled"`
	CancellationRate float64 `json:"cancellationRate"`
	ScheduledMinutes int64   `json:"scheduledMinutes"`
	ConductedMinutes int64   `json:"conductedMinutes"`
	UtilizationPct   float64 `json:"utilizationPct"`
	NoShowCount      int64   `json:"noShowCount"`
}

type analyticsStudentRowJSON struct {
	StudentID        string  `json:"studentId"`
	StudentName      string  `json:"studentName"`
	Conducted        int64   `json:"conducted"`
	Cancelled        int64   `json:"cancelled"`
	Rescheduled      int64   `json:"rescheduled"`
	CancellationRate float64 `json:"cancellationRate"`
}

type analyticsWeeklyRowJSON struct {
	WeekLabel        string  `json:"weekLabel"`
	Conducted        int64   `json:"conducted"`
	Cancelled        int64   `json:"cancelled"`
	Rescheduled      int64   `json:"rescheduled"`
	CancellationRate float64 `json:"cancellationRate"`
}

type analyticsNoShowJSON struct {
	ID              string            `json:"id"`
	ScheduledDate   string            `json:"scheduledDate"`
	StartTime       string            `json:"startTime"`
	DurationMinutes int64             `json:"durationMinutes"`
	StudentName     string            `json:"studentName"`
	TeacherName     string            `json:"teacherName"`
	TeacherAvatar   models.AvatarView `json:"teacherAvatar"`
	DaysOverdue     int64             `json:"daysOverdue"`
}

type analyticsInactiveReasonJSON struct {
	Reason string `json:"reason"`
	Count  int64  `json:"count"`
}

type analyticsChurnedStudentJSON struct {
	StudentID      string `json:"studentId"`
	StudentName    string `json:"studentName"`
	InactiveReason string `json:"inactiveReason"`
	TenureDays     int64  `json:"tenureDays"`
	ChurnedAt      string `json:"churnedAt"`
}

type analyticsRetentionJSON struct {
	ActiveCount      int64 `json:"activeCount"`
	InactiveCount    int64 `json:"inactiveCount"`
	ChurnedInPeriod  int64 `json:"churnedInPeriod"`
	MedianTenureDays int64 `json:"medianTenureDays"`
}

type analyticsResponseJSON struct {
	Summary    analyticsSummaryJSON        `json:"summary"`
	ByTeacher  []analyticsTeacherRowJSON   `json:"byTeacher"`
	ByStudent  []analyticsStudentRowJSON   `json:"byStudent"`
	Weekly     []analyticsWeeklyRowJSON    `json:"weekly"`
	NoShows    []analyticsNoShowJSON       `json:"noShows"`
	Retention        analyticsRetentionJSON        `json:"retention"`
	InactiveReasons  []analyticsInactiveReasonJSON `json:"inactiveReasons"`
	ChurnedStudents  []analyticsChurnedStudentJSON `json:"churnedStudents"`
}

func handleAnalytics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := frontend.AnalyticsData{}
	role := auth.GetRole(r.Context())
	if auth.HasAdminAccess(role) {
		data.IsSuperuser = true
	} else {
		user := auth.GetUser(r.Context())
		data.TeacherID = strconv.FormatInt(user.ID, 10)
		data.TeacherName = user.Name
	}

	w.Header().Set("Content-Type", "text/html")
	if err := frontend.Analytics(data).Render(r.Context(), w); err != nil {
		logs.Log().Error("render analytics", zap.Error(err))
	}
}

func handleGetAnalytics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	startDate, endDate, err := requireReportDateRange(r)
	if err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	teacherID, err := analyticsTeacherID(r)
	if err != nil {
		if err.Error() == "forbidden" {
			HttpError(w, "Forbidden", http.StatusForbidden)
			return
		}
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	q := dbRO.GetQueries()
	isSuperuser := auth.HasAdminAccess(auth.GetRole(ctx))

	summaryRow, err := q.GetAnalyticsSummary(ctx, analyticsSummaryParams(startDate, endDate, teacherID))
	if err != nil {
		logs.Log().Error("analytics summary", zap.Error(err))
		HttpError(w, "Failed to load analytics", http.StatusInternalServerError)
		return
	}

	conducted := sqlNumericToInt64(summaryRow.Conducted)
	cancelled := sqlNumericToInt64(summaryRow.Cancelled)
	rescheduled := sqlNumericToInt64(summaryRow.Rescheduled)
	scheduledMinutes := sqlNumericToInt64(summaryRow.ScheduledMinutes)
	conductedMinutes := sqlNumericToInt64(summaryRow.ConductedMinutes)
	noShowCount := sqlNumericToInt64(summaryRow.NoShowCount)

	resp := analyticsResponseJSON{
		Summary: analyticsSummaryJSON{
			Conducted:        conducted,
			Cancelled:        cancelled,
			Rescheduled:      rescheduled,
			CancellationRate: cancellationRate(cancelled, conducted, rescheduled),
			ScheduledMinutes: scheduledMinutes,
			ConductedMinutes: conductedMinutes,
			UtilizationPct:   utilizationPct(conductedMinutes, scheduledMinutes),
			NoShowCount:      noShowCount,
		},
	}

	if isSuperuser {
		teacherRows, err := q.GetAnalyticsCancellationByTeacher(ctx, analyticsByTeacherParams(startDate, endDate, teacherID))
		if err != nil {
			logs.Log().Error("analytics by teacher", zap.Error(err))
			HttpError(w, "Failed to load analytics", http.StatusInternalServerError)
			return
		}
		resp.ByTeacher = make([]analyticsTeacherRowJSON, 0, len(teacherRows))
		for _, row := range teacherRows {
			tConducted := sqlNumericToInt64(row.Conducted)
			tCancelled := sqlNumericToInt64(row.Cancelled)
			tRescheduled := sqlNumericToInt64(row.Rescheduled)
			tScheduledMinutes := sqlNumericToInt64(row.ScheduledMinutes)
			tConductedMinutes := sqlNumericToInt64(row.ConductedMinutes)
			resp.ByTeacher = append(resp.ByTeacher, analyticsTeacherRowJSON{
				TeacherID:        strconv.FormatInt(row.TeacherID, 10),
				TeacherName:      row.TeacherName,
				TeacherAvatar:    analyticsTeacherAvatar(row.TeacherID, row.TeacherFirstName, row.TeacherMiddleName, row.TeacherLastName, row.TeacherName, row.TeacherAssignedColor, row.TeacherProfilePicture),
				Conducted:        tConducted,
				Cancelled:        tCancelled,
				Rescheduled:      tRescheduled,
				CancellationRate: cancellationRate(tCancelled, tConducted, tRescheduled),
				ScheduledMinutes: tScheduledMinutes,
				ConductedMinutes: tConductedMinutes,
				UtilizationPct:   utilizationPct(tConductedMinutes, tScheduledMinutes),
				NoShowCount:      sqlNumericToInt64(row.NoShowCount),
			})
		}
	}

	studentRows, err := q.GetAnalyticsCancellationByStudent(ctx, analyticsByStudentParams(startDate, endDate, teacherID))
	if err != nil {
		logs.Log().Error("analytics by student", zap.Error(err))
		HttpError(w, "Failed to load analytics", http.StatusInternalServerError)
		return
	}
	resp.ByStudent = make([]analyticsStudentRowJSON, 0, len(studentRows))
	for _, row := range studentRows {
		sConducted := sqlNumericToInt64(row.Conducted)
		sCancelled := sqlNumericToInt64(row.Cancelled)
		sRescheduled := sqlNumericToInt64(row.Rescheduled)
		resp.ByStudent = append(resp.ByStudent, analyticsStudentRowJSON{
			StudentID:        strconv.FormatInt(row.StudentID, 10),
			StudentName:      row.StudentName,
			Conducted:        sConducted,
			Cancelled:        sCancelled,
			Rescheduled:      sRescheduled,
			CancellationRate: cancellationRate(sCancelled, sConducted, sRescheduled),
		})
	}

	weeklyRows, err := q.GetAnalyticsWeeklyTrend(ctx, queries.GetAnalyticsWeeklyTrendParams{
		Date:      startDate,
		Date_2:    endDate,
		Column3:   int64(0),
		TeacherID: teacherID,
	})
	if err != nil {
		logs.Log().Error("analytics weekly", zap.Error(err))
		HttpError(w, "Failed to load analytics", http.StatusInternalServerError)
		return
	}
	resp.Weekly = make([]analyticsWeeklyRowJSON, 0, len(weeklyRows))
	for _, row := range weeklyRows {
		wConducted := sqlNumericToInt64(row.Conducted)
		wCancelled := sqlNumericToInt64(row.Cancelled)
		wRescheduled := sqlNumericToInt64(row.Rescheduled)
		weekLabel := ""
		if row.WeekLabel != nil {
			weekLabel = sqlNumericToString(row.WeekLabel)
		}
		resp.Weekly = append(resp.Weekly, analyticsWeeklyRowJSON{
			WeekLabel:        weekLabel,
			Conducted:        wConducted,
			Cancelled:        wCancelled,
			Rescheduled:      wRescheduled,
			CancellationRate: cancellationRate(wCancelled, wConducted, wRescheduled),
		})
	}

	noShowRows, err := q.GetAnalyticsNoShows(ctx, queries.GetAnalyticsNoShowsParams{
		ScheduledDate:   startDate,
		ScheduledDate_2: endDate,
		Column3:         int64(0),
		TeacherID:       teacherID,
	})
	if err != nil {
		logs.Log().Error("analytics no-shows", zap.Error(err))
		HttpError(w, "Failed to load analytics", http.StatusInternalServerError)
		return
	}
	resp.NoShows = make([]analyticsNoShowJSON, 0, len(noShowRows))
	noShowTeacherIDs := make([]int64, 0, len(noShowRows))
	for _, row := range noShowRows {
		noShowTeacherIDs = append(noShowTeacherIDs, row.TeacherID)
		resp.NoShows = append(resp.NoShows, analyticsNoShowJSON{
			ID:              strconv.FormatInt(row.ID, 10),
			ScheduledDate:   row.ScheduledDate,
			StartTime:       row.StartTime.String,
			DurationMinutes: row.DurationMinutes,
			StudentName:     row.StudentName,
			TeacherName:     row.TeacherName,
			TeacherAvatar:   analyticsTeacherAvatar(row.TeacherID, row.TeacherFirstName, row.TeacherMiddleName, row.TeacherLastName, row.TeacherName, row.TeacherAssignedColor, row.TeacherProfilePicture),
			DaysOverdue:     row.DaysOverdue,
		})
	}

	inactiveReasonRows, err := q.GetAnalyticsInactiveReasons(ctx, queries.GetAnalyticsInactiveReasonsParams{
		Column1:   int64(0),
		TeacherID: teacherID,
	})
	if err != nil {
		logs.Log().Error("analytics inactive reasons", zap.Error(err))
		HttpError(w, "Failed to load analytics", http.StatusInternalServerError)
		return
	}
	resp.InactiveReasons = make([]analyticsInactiveReasonJSON, 0, len(inactiveReasonRows))
	for _, row := range inactiveReasonRows {
		resp.InactiveReasons = append(resp.InactiveReasons, analyticsInactiveReasonJSON{
			Reason: sqlNumericToString(row.InactiveReason),
			Count:  row.Count,
		})
	}

	churnedRows, err := q.GetAnalyticsChurnedStudents(ctx, queries.GetAnalyticsChurnedStudentsParams{
		UpdatedAt:   analyticsDateNullTime(startDate),
		UpdatedAt_2: analyticsDateNullTime(endDate),
		Column3:     int64(0),
		TeacherID:   teacherID,
	})
	if err != nil {
		logs.Log().Error("analytics churned students", zap.Error(err))
		HttpError(w, "Failed to load analytics", http.StatusInternalServerError)
		return
	}

	tenures := make([]int64, 0, len(churnedRows))
	resp.ChurnedStudents = make([]analyticsChurnedStudentJSON, 0, len(churnedRows))
	for _, row := range churnedRows {
		tenures = append(tenures, row.TenureDays)
		resp.ChurnedStudents = append(resp.ChurnedStudents, analyticsChurnedStudentJSON{
			StudentID:      strconv.FormatInt(row.StudentID, 10),
			StudentName:    row.StudentName,
			InactiveReason: row.InactiveReason,
			TenureDays:     row.TenureDays,
			ChurnedAt:      utils.FormatNullDateTimePHT(row.ChurnedAt),
		})
	}

	retentionRow, err := q.GetAnalyticsRetentionSummary(ctx, analyticsRetentionParams(startDate, endDate, teacherID))
	if err != nil {
		logs.Log().Error("analytics retention summary", zap.Error(err))
		HttpError(w, "Failed to load analytics", http.StatusInternalServerError)
		return
	}
	resp.Retention = analyticsRetentionJSON{
		ActiveCount:      sqlNumericToInt64(retentionRow.ActiveCount),
		InactiveCount:    sqlNumericToInt64(retentionRow.InactiveCount),
		ChurnedInPeriod:  sqlNumericToInt64(retentionRow.ChurnedInPeriod),
		MedianTenureDays: medianInt64(tenures),
	}

	if err := enrichAnalyticsResponseWithRoleBadges(ctx, &resp, noShowTeacherIDs); err != nil {
		logs.Log().Error("analytics teacher roles", zap.Error(err))
		HttpError(w, "Failed to load analytics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logs.Log().Error("encode analytics", zap.Error(err))
	}
}

func enrichAnalyticsResponseWithRoleBadges(ctx context.Context, resp *analyticsResponseJSON, noShowTeacherIDs []int64) error {
	teacherIDs := make([]int64, 0, len(resp.ByTeacher)+len(noShowTeacherIDs))
	for _, row := range resp.ByTeacher {
		id, err := strconv.ParseInt(row.TeacherID, 10, 64)
		if err != nil {
			continue
		}
		teacherIDs = append(teacherIDs, id)
	}
	teacherIDs = append(teacherIDs, noShowTeacherIDs...)

	rolesMap, err := loadRolesByTeacherIDs(ctx, uniqueTeacherIDs(teacherIDs))
	if err != nil {
		return err
	}

	for i := range resp.ByTeacher {
		id, err := strconv.ParseInt(resp.ByTeacher[i].TeacherID, 10, 64)
		if err != nil {
			continue
		}
		resp.ByTeacher[i].TeacherAvatar = avatarViewWithTeacherRoles(resp.ByTeacher[i].TeacherAvatar, rolesMap[id])
	}
	for i, teacherID := range noShowTeacherIDs {
		resp.NoShows[i].TeacherAvatar = avatarViewWithTeacherRoles(resp.NoShows[i].TeacherAvatar, rolesMap[teacherID])
	}
	return nil
}

func analyticsTeacherID(r *http.Request) (int64, error) {
	role := auth.GetRole(r.Context())
	if role == auth.RoleTeacher {
		user := auth.GetUser(r.Context())
		if user.ID == 0 {
			return 0, errors.New("unauthorized")
		}
		if teacherIDStr := strings.TrimSpace(r.URL.Query().Get("teacherId")); teacherIDStr != "" {
			parsedID, err := strconv.ParseInt(teacherIDStr, 10, 64)
			if err != nil {
				return 0, errors.New("invalid teacher ID")
			}
			if parsedID != user.ID {
				return 0, errors.New("forbidden")
			}
		}
		return user.ID, nil
	}

	teacherIDStr := strings.TrimSpace(r.URL.Query().Get("teacherId"))
	if teacherIDStr == "" {
		return 0, nil
	}
	parsedID, err := strconv.ParseInt(teacherIDStr, 10, 64)
	if err != nil {
		return 0, errors.New("invalid teacher ID")
	}
	return parsedID, nil
}

func analyticsSummaryParams(startDate, endDate string, teacherID int64) queries.GetAnalyticsSummaryParams {
	return queries.GetAnalyticsSummaryParams{
		Date:            startDate,
		Date_2:          endDate,
		Column3:         int64(0),
		TeacherID:       teacherID,
		Date_3:          startDate,
		Date_4:          endDate,
		Column7:         int64(0),
		TeacherID_2:     teacherID,
		Date_5:          startDate,
		Date_6:          endDate,
		Column11:        int64(0),
		TeacherID_3:     teacherID,
		ScheduledDate:   startDate,
		ScheduledDate_2: endDate,
		Column15:        int64(0),
		TeacherID_4:     teacherID,
		Date_7:          startDate,
		Date_8:          endDate,
		Column19:        int64(0),
		TeacherID_5:     teacherID,
		ScheduledDate_3: startDate,
		ScheduledDate_4: endDate,
		Column23:        int64(0),
		TeacherID_6:     teacherID,
	}
}

func analyticsByTeacherParams(startDate, endDate string, teacherID int64) queries.GetAnalyticsCancellationByTeacherParams {
	return queries.GetAnalyticsCancellationByTeacherParams{
		ScheduledDate:   startDate,
		ScheduledDate_2: endDate,
		Date:            startDate,
		Date_2:          endDate,
		ScheduledDate_3: startDate,
		ScheduledDate_4: endDate,
		Date_3:          startDate,
		Date_4:          endDate,
		Column9:         int64(0),
		ID:              teacherID,
	}
}

func analyticsByStudentParams(startDate, endDate string, teacherID int64) queries.GetAnalyticsCancellationByStudentParams {
	return queries.GetAnalyticsCancellationByStudentParams{
		Date:        startDate,
		Date_2:      endDate,
		Column3:     int64(0),
		TeacherID:   teacherID,
		Column5:     int64(0),
		TeacherID_2: teacherID,
	}
}

func analyticsRetentionParams(startDate, endDate string, teacherID int64) queries.GetAnalyticsRetentionSummaryParams {
	return queries.GetAnalyticsRetentionSummaryParams{
		Column1:     int64(0),
		TeacherID:   teacherID,
		Column3:     int64(0),
		TeacherID_2: teacherID,
		UpdatedAt:   analyticsDateNullTime(startDate),
		UpdatedAt_2: analyticsDateNullTime(endDate),
		Column7:     int64(0),
		TeacherID_3: teacherID,
	}
}

func analyticsDateNullTime(date string) sql.NullTime {
	t, err := utils.ParseDatePHT(date)
	if err != nil || t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func cancellationRate(cancelled, conducted, rescheduled int64) float64 {
	total := conducted + cancelled + rescheduled
	if total == 0 {
		return 0
	}
	return math.Round(float64(cancelled)/float64(total)*10000) / 100
}

func utilizationPct(conductedMinutes, scheduledMinutes int64) float64 {
	if scheduledMinutes == 0 {
		return 0
	}
	return math.Round(float64(conductedMinutes)/float64(scheduledMinutes)*10000) / 100
}

func medianInt64(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]int64(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

func analyticsTeacherAvatar(teacherID int64, firstName, middleName, lastName, teacherName, assignedColor string, profilePicture sql.NullString) models.AvatarView {
	hasPicture := profilePicture.Valid && profilePicture.String != ""
	if assignedColor == "" {
		assignedColor = constants.DefaultTeacherAssignedColor
	}
	return models.AvatarView{
		Initials:      utils.PersonInitials(firstName, middleName, lastName, teacherName),
		AssignedColor: assignedColor,
		HasPicture:    hasPicture,
		PictureURL:    teacherPictureURL(teacherID, hasPicture),
		Alt:           teacherName + " avatar",
	}
}

func sqlNumericToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return ""
	}
}
