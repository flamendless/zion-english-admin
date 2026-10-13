package utils

import (
	"database/sql"
	"strings"
)

func StudentDisplayName(name, studentNumber string) string {
	name = strings.TrimSpace(name)
	num := strings.TrimSpace(studentNumber)
	if num == "" {
		return name
	}
	return name + " (" + num + ")"
}

func StudentDisplayNameFromNull(name string, studentNumber sql.NullString) string {
	if !studentNumber.Valid {
		return name
	}
	return StudentDisplayName(name, studentNumber.String)
}
