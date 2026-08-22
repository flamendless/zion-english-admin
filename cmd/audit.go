package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"zion-english/internal/auth"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/utils"
	"go.uber.org/zap"
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

func auditStr(v sql.NullString) string {
	if v.Valid && v.String != "" {
		return v.String
	}
	return "(empty)"
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
		parts = append(parts, fmt.Sprintf("student id %d -> %d", before.StudentID, after.StudentID))
	}
	if before.TeacherID != after.TeacherID {
		parts = append(parts, fmt.Sprintf("teacher id %d -> %d", before.TeacherID, after.TeacherID))
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
	if before.Status != after.Status {
		parts = append(parts, fmt.Sprintf("status '%s' -> '%s'", before.Status, after.Status))
	}
	if before.InactiveReason.String != after.InactiveReason.String {
		parts = append(parts, fmt.Sprintf("inactive reason '%s' -> '%s'", auditStr(before.InactiveReason), auditStr(after.InactiveReason)))
	}
	if before.RatePerClass != after.RatePerClass {
		parts = append(parts, fmt.Sprintf("rate %.2f -> %.2f", before.RatePerClass, after.RatePerClass))
	}
	if before.Currency != after.Currency {
		parts = append(parts, fmt.Sprintf("currency '%s' -> '%s'", before.Currency, after.Currency))
	}
	if teachersBefore != teachersAfter {
		parts = append(parts, fmt.Sprintf("teachers [%s] -> [%s]", teachersBefore, teachersAfter))
	}
	if len(parts) == 0 {
		return fmt.Sprintf("updated student '%s' (no field changes)", after.Name)
	}
	return fmt.Sprintf("updated student '%s': %s", after.Name, strings.Join(parts, ", "))
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
