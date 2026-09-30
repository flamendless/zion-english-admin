package utils

import (
	"database/sql"
	"errors"

	"github.com/mattn/go-sqlite3"
)

func NullStringFromAny(v any) sql.NullString {
	s := InterfaceToString(v)
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func IsUniqueConstraint(err error) bool {
	var sqliteErr sqlite3.Error
	if !errors.As(err, &sqliteErr) {
		return false
	}
	return sqliteErr.Code == sqlite3.ErrConstraint && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique
}
