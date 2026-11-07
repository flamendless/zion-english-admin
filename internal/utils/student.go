package utils

import (
	"database/sql"
	"strings"
)

func StudentDisplayName(name, studentNumber, legalName string) string {
	name = strings.TrimSpace(name)
	num := strings.TrimSpace(studentNumber)
	legalName = strings.TrimSpace(legalName)

	base := name
	if num != "" {
		base = name + " (" + num + ")"
	}
	if legalName != "" {
		return base + " (" + legalName + ")"
	}
	return base
}

func StudentDisplayNameFromNull(name string, studentNumber, legalName sql.NullString) string {
	num := ""
	if studentNumber.Valid {
		num = studentNumber.String
	}
	legal := ""
	if legalName.Valid {
		legal = legalName.String
	}
	return StudentDisplayName(name, num, legal)
}
