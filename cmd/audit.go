package cmd

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"zion-english/internal/database/queries"
)

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

func formatStudentAudit(before queries.TblStudent, after queries.TblStudent, teachersBefore, teachersAfter string) string {
	var parts []string
	if before.Name != after.Name {
		parts = append(parts, fmt.Sprintf("name '%s' -> '%s'", before.Name, after.Name))
	}
	if before.Status != after.Status {
		parts = append(parts, fmt.Sprintf("status '%s' -> '%s'", before.Status, after.Status))
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
