package cmd

import (
	"database/sql"

	"zion-english/internal/utils"
)

func studentDisplayName(name string, studentNumber sql.NullString) string {
	return utils.StudentDisplayNameFromNull(name, studentNumber)
}

func studentNumberString(studentNumber sql.NullString) string {
	if !studentNumber.Valid {
		return ""
	}
	return studentNumber.String
}

func studentIdentifierFieldsFromForm(studentNumberRaw, parentNumberRaw string) (sql.NullString, sql.NullString, error) {
	studentNumber, err := parseOptionalStudentIdentifier(studentNumberRaw, ErrInvalidStudentNumber)
	if err != nil {
		return sql.NullString{}, sql.NullString{}, err
	}
	parentNumber, err := parseOptionalStudentIdentifier(parentNumberRaw, ErrInvalidParentNumber)
	if err != nil {
		return sql.NullString{}, sql.NullString{}, err
	}
	return studentNumber, parentNumber, nil
}
