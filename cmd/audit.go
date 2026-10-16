package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"zion-english/internal/auth"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/utils"
)

func insertAuditLogAs(ctx context.Context, actor auth.User, module, message string) {
	var createdBy sql.NullInt64
	if actor.ID > 0 {
		createdBy = sql.NullInt64{Int64: actor.ID, Valid: true}
	}
	if err := dbRW.GetQueries().InsertLog(ctx, queries.InsertLogParams{
		Module:        module,
		Message:       message,
		CreatedBy:     createdBy,
		CreatedByName: sql.NullString{String: actor.Name, Valid: actor.Name != ""},
	}); err != nil {
		logs.Log().Info("system logs", zap.Error(err))
	}
}

type uploadLogEntry struct {
	Module         string
	Outcome        constants.UploadLogOutcome
	Kind           constants.UploadLogKind
	Summary        string
	Filename       string
	FileSize       int64
	FileSizeValid  bool
	CompressPreset string
}

func insertUploadLog(ctx context.Context, actor auth.User, entry uploadLogEntry) {
	var createdBy sql.NullInt64
	if actor.ID > 0 {
		createdBy = sql.NullInt64{Int64: actor.ID, Valid: true}
	}
	var filename sql.NullString
	if entry.Filename != "" {
		filename = sql.NullString{String: entry.Filename, Valid: true}
	}
	var fileSize sql.NullInt64
	if entry.FileSizeValid {
		fileSize = sql.NullInt64{Int64: entry.FileSize, Valid: true}
	}
	var compressPreset sql.NullString
	if entry.CompressPreset != "" {
		compressPreset = sql.NullString{String: entry.CompressPreset, Valid: true}
	}
	if err := dbRW.GetQueries().InsertUploadLog(ctx, queries.InsertUploadLogParams{
		Module:         entry.Module,
		Outcome:        string(entry.Outcome),
		Kind:           string(entry.Kind),
		Summary:        entry.Summary,
		Filename:       filename,
		FileSize:       fileSize,
		CompressPreset: compressPreset,
		CreatedBy:      createdBy,
		CreatedByName:  sql.NullString{String: actor.Name, Valid: actor.Name != ""},
	}); err != nil {
		logs.Log().Info("upload logs", zap.Error(err))
	}
}

func auditStr(v sql.NullString) string {
	if v.Valid && v.String != "" {
		return v.String
	}
	return "(empty)"
}

func auditStudentRef(id int64, name string, studentNumber sql.NullString) string {
	return fmt.Sprintf("student id %d (%s)", id, studentDisplayName(name, studentNumber))
}

func auditTeacherRef(id int64, name string) string {
	label := strings.TrimSpace(name)
	if label == "" {
		label = "unknown"
	}
	return fmt.Sprintf("teacher id %d (%s)", id, label)
}

func auditStudentTeacherPair(studentID int64, studentName string, studentNumber sql.NullString, teacherID int64, teacherName string) string {
	return auditStudentRef(studentID, studentName, studentNumber) + ", " + auditTeacherRef(teacherID, teacherName)
}

func auditPartiesFromScheduled(row queries.GetScheduledClassByIDRow) string {
	return auditStudentTeacherPair(row.StudentID, row.StudentName, row.StudentNumber, row.TeacherID, row.TeacherName)
}

func auditPartiesFromClassRecord(row queries.GetClassRecordByIDRow) string {
	return auditStudentTeacherPair(row.StudentID, row.StudentName, row.StudentNumber, row.TeacherID, row.TeacherName)
}

func auditStudentRefCtx(ctx context.Context, id int64) string {
	s, err := dbRO.GetQueries().GetStudentByID(ctx, id)
	if err != nil {
		return fmt.Sprintf("student id %d (unknown)", id)
	}
	return auditStudentRef(id, s.Name, s.StudentNumber)
}

func auditTeacherRefCtx(ctx context.Context, id int64) string {
	return auditTeacherRef(id, teacherNameByID(ctx, id))
}

func auditTeacherRefsList(ctx context.Context, ids []int64) string {
	if len(ids) == 0 {
		return ""
	}
	names := teacherNamesMap(ctx, ids)
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = auditTeacherRef(id, names[id])
	}
	return strings.Join(parts, ", ")
}

func auditStudentRefForEdit(ctx context.Context, studentID int64, existing queries.GetScheduledClassByIDRow) string {
	if studentID == existing.StudentID {
		return auditStudentRef(existing.StudentID, existing.StudentName, existing.StudentNumber)
	}
	return auditStudentRefCtx(ctx, studentID)
}

func formatRecordedClassAudit(parties string, date string, status string) string {
	return fmt.Sprintf("recorded class for %s, date %s, status %s", parties, date, status)
}

func formatClassRecordAudit(before, after queries.GetClassRecordByIDRow) string {
	var parts []string
	if before.Date != after.Date {
		parts = append(parts, fmt.Sprintf("date '%s' -> '%s'", before.Date, after.Date))
	}
	if before.DurationMinutes != after.DurationMinutes {
		parts = append(parts, fmt.Sprintf("duration %d -> %d", before.DurationMinutes, after.DurationMinutes))
	}
	if before.Rate != after.Rate {
		parts = append(parts, fmt.Sprintf("rate %.2f -> %.2f", before.Rate, after.Rate))
	}
	if before.Currency != after.Currency {
		parts = append(parts, fmt.Sprintf("currency '%s' -> '%s'", before.Currency, after.Currency))
	}
	if before.Status != after.Status {
		parts = append(parts, fmt.Sprintf("status '%s' -> '%s'", before.Status, after.Status))
	}
	if before.Reason.String != after.Reason.String {
		parts = append(parts, fmt.Sprintf("reason '%s' -> '%s'", auditStr(before.Reason), auditStr(after.Reason)))
	}
	if before.Notes.String != after.Notes.String {
		parts = append(parts, fmt.Sprintf("notes '%s' -> '%s'", auditStr(before.Notes), auditStr(after.Notes)))
	}
	if before.StudentID != after.StudentID {
		parts = append(parts, fmt.Sprintf("%s -> %s",
			auditStudentRef(before.StudentID, before.StudentName, before.StudentNumber),
			auditStudentRef(after.StudentID, after.StudentName, after.StudentNumber)))
	}
	if before.TeacherID != after.TeacherID {
		parts = append(parts, fmt.Sprintf("%s -> %s",
			auditTeacherRef(before.TeacherID, before.TeacherName),
			auditTeacherRef(after.TeacherID, after.TeacherName)))
	}
	if len(parts) == 0 {
		return fmt.Sprintf("updated class record id %d (no field changes)", after.ID)
	}
	return fmt.Sprintf("updated class record id %d: %s", after.ID, strings.Join(parts, ", "))
}

func formatTeacherAudit(before, after queries.GetTeacherFullByIDRow) string {
	var parts []string
	if before.FirstName != after.FirstName {
		parts = append(parts, fmt.Sprintf("first name '%s' -> '%s'", before.FirstName, after.FirstName))
	}
	if before.MiddleName != after.MiddleName {
		parts = append(parts, fmt.Sprintf("middle name '%s' -> '%s'", before.MiddleName, after.MiddleName))
	}
	if before.LastName != after.LastName {
		parts = append(parts, fmt.Sprintf("last name '%s' -> '%s'", before.LastName, after.LastName))
	}
	if before.Birthdate != after.Birthdate {
		parts = append(parts, fmt.Sprintf("birthdate '%s' -> '%s'", before.Birthdate, after.Birthdate))
	}
	if before.Address != after.Address {
		parts = append(parts, fmt.Sprintf("address '%s' -> '%s'", before.Address, after.Address))
	}
	if before.JoiningDate != after.JoiningDate {
		parts = append(parts, fmt.Sprintf("joining date '%s' -> '%s'", before.JoiningDate, after.JoiningDate))
	}
	if before.MobileNumber != after.MobileNumber {
		parts = append(parts, fmt.Sprintf("mobile '%s' -> '%s'", before.MobileNumber, after.MobileNumber))
	}
	if before.Email != after.Email {
		parts = append(parts, fmt.Sprintf("email '%s' -> '%s'", before.Email, after.Email))
	}
	if before.Certifications.String != after.Certifications.String {
		parts = append(parts, fmt.Sprintf("certifications '%s' -> '%s'", auditStr(before.Certifications), auditStr(after.Certifications)))
	}
	if before.AssignedColor != after.AssignedColor {
		parts = append(parts, fmt.Sprintf("assigned color '%s' -> '%s'", before.AssignedColor, after.AssignedColor))
	}
	if before.RatePerClass != after.RatePerClass {
		parts = append(parts, fmt.Sprintf("rate %.2f -> %.2f", before.RatePerClass, after.RatePerClass))
	}
	if before.RateDurationMinutes != after.RateDurationMinutes {
		parts = append(parts, fmt.Sprintf("rate duration %d -> %d min", before.RateDurationMinutes, after.RateDurationMinutes))
	}
	if before.Currency != after.Currency {
		parts = append(parts, fmt.Sprintf("currency '%s' -> '%s'", before.Currency, after.Currency))
	}
	if before.DriveUrl != after.DriveUrl {
		parts = append(parts, fmt.Sprintf("drive url '%s' -> '%s'", before.DriveUrl, after.DriveUrl))
	}
	if before.Sex.String != after.Sex.String {
		parts = append(parts, fmt.Sprintf("sex '%s' -> '%s'", auditStr(before.Sex), auditStr(after.Sex)))
	}
	if before.Template.String != after.Template.String {
		parts = append(parts, fmt.Sprintf("template '%s' -> '%s'", auditStr(before.Template), auditStr(after.Template)))
	}
	if before.Status != after.Status {
		parts = append(parts, fmt.Sprintf("status '%s' -> '%s'", before.Status, after.Status))
	}
	teacherName := utils.ComposePersonName(after.FirstName, after.MiddleName, after.LastName)
	if len(parts) == 0 {
		return fmt.Sprintf("updated teacher '%s' (id %d) (no field changes)", teacherName, after.ID)
	}
	return fmt.Sprintf("updated teacher '%s' (id %d): %s", teacherName, after.ID, strings.Join(parts, ", "))
}

func formatStudentAudit(before, after queries.GetStudentByIDRow, teachersBefore, teachersAfter string) string {
	var parts []string
	if before.Name != after.Name {
		parts = append(parts, fmt.Sprintf("name '%s' -> '%s'", before.Name, after.Name))
	}
	if before.StudentNumber.String != after.StudentNumber.String || before.StudentNumber.Valid != after.StudentNumber.Valid {
		parts = append(parts, fmt.Sprintf("student number '%s' -> '%s'", auditStr(before.StudentNumber), auditStr(after.StudentNumber)))
	}
	if before.ParentNumber.String != after.ParentNumber.String || before.ParentNumber.Valid != after.ParentNumber.Valid {
		parts = append(parts, fmt.Sprintf("parent number '%s' -> '%s'", auditStr(before.ParentNumber), auditStr(after.ParentNumber)))
	}
	if before.Status != after.Status {
		parts = append(parts, fmt.Sprintf("status '%s' -> '%s'", before.Status, after.Status))
	}
	if before.InactiveReason.String != after.InactiveReason.String {
		parts = append(parts, fmt.Sprintf("inactive reason '%s' -> '%s'", auditStr(before.InactiveReason), auditStr(after.InactiveReason)))
	}
	if before.RatePerClass != after.RatePerClass {
		parts = append(parts, fmt.Sprintf("rate %.2f -> %.2f", before.RatePerClass, after.RatePerClass))
	}
	if before.RateDurationMinutes != after.RateDurationMinutes {
		parts = append(parts, fmt.Sprintf("rate duration %d -> %d min", before.RateDurationMinutes, after.RateDurationMinutes))
	}
	if before.Currency != after.Currency {
		parts = append(parts, fmt.Sprintf("currency '%s' -> '%s'", before.Currency, after.Currency))
	}
	if before.ParentName.String != after.ParentName.String {
		parts = append(parts, fmt.Sprintf("parent name '%s' -> '%s'", auditStr(before.ParentName), auditStr(after.ParentName)))
	}
	if before.ParentRate.Float64 != after.ParentRate.Float64 || before.ParentRate.Valid != after.ParentRate.Valid {
		parts = append(parts, fmt.Sprintf("parent rate '%s' -> '%s'", auditParentRate(before.ParentRate, before.ParentCurrency), auditParentRate(after.ParentRate, after.ParentCurrency)))
	}
	if before.AssignedColor != after.AssignedColor {
		parts = append(parts, fmt.Sprintf("assigned color '%s' -> '%s'", before.AssignedColor, after.AssignedColor))
	}
	if teachersBefore != teachersAfter {
		parts = append(parts, fmt.Sprintf("teachers [%s] -> [%s]", teachersBefore, teachersAfter))
	}
	if len(parts) == 0 {
		return fmt.Sprintf("updated student '%s' (no field changes)", after.Name)
	}
	return fmt.Sprintf("updated student '%s': %s", after.Name, strings.Join(parts, ", "))
}

func auditParentRate(rate sql.NullFloat64, currency sql.NullString) string {
	if !rate.Valid || !currency.Valid {
		return ""
	}
	return fmt.Sprintf("%.2f %s", rate.Float64, currency.String)
}

func teacherIDsString(teachers []queries.GetTeachersByStudentIDRow) string {
	if len(teachers) == 0 {
		return ""
	}
	ids := make([]string, len(teachers))
	for i, t := range teachers {
		ids[i] = strconv.FormatInt(t.ID, 10)
	}
	return strings.Join(ids, ",")
}
