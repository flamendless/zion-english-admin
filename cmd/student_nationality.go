package cmd

import (
	"context"
	"database/sql"
	"strings"
	"unicode/utf8"

	"zion-english/frontend"
	"zion-english/internal/database/queries"
	"zion-english/internal/utils"
)

const studentLegalNameMaxRunes = 64

func nationalityFlagMap(ctx context.Context) (map[string]string, error) {
	rows, err := dbRO.GetQueries().GetAllNationalities(ctx)
	if err != nil {
		return nil, err
	}
	flags := make(map[string]string, len(rows))
	for _, row := range rows {
		flags[row.ID] = utils.URL(row.FlagPath)
	}
	return flags, nil
}

func loadNationalityOptions(ctx context.Context) ([]frontend.NationalityOption, error) {
	rows, err := dbRO.GetQueries().GetAllNationalities(ctx)
	if err != nil {
		return nil, err
	}
	options := make([]frontend.NationalityOption, len(rows))
	for i, row := range rows {
		lang := ""
		if row.Language.Valid {
			lang = row.Language.String
		}
		options[i] = frontend.NationalityOption{
			ID:          row.ID,
			CountryName: row.CountryName,
			Language:    lang,
			FlagURL:     utils.URL(row.FlagPath),
		}
	}
	return options, nil
}

func studentLegalNameString(v sql.NullString) string {
	if !v.Valid {
		return ""
	}
	return strings.TrimSpace(v.String)
}

func studentNationalityIDString(v sql.NullString) string {
	if !v.Valid {
		return ""
	}
	return strings.TrimSpace(v.String)
}

func parseStudentLegalName(raw string) (sql.NullString, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return sql.NullString{}, nil
	}
	if utf8.RuneCountInString(raw) > studentLegalNameMaxRunes {
		return sql.NullString{}, ErrLegalNameTooLong
	}
	return sql.NullString{String: raw, Valid: true}, nil
}

func parseStudentNationalityID(ctx context.Context, raw string) (sql.NullString, error) {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" {
		return sql.NullString{}, nil
	}
	_, err := dbRO.GetQueries().GetNationalityByID(ctx, raw)
	if err != nil {
		return sql.NullString{}, ErrInvalidNationality
	}
	return sql.NullString{String: raw, Valid: true}, nil
}

func buildStudentAvatarProps(name, assignedColor string, isTestSeed bool, size string, nationalityID sql.NullString, flags map[string]string) frontend.AvatarProps {
	props := frontend.StudentAvatarProps(name, assignedColor, isTestSeed, size)
	if nationalityID.Valid && flags != nil {
		if flagURL, ok := flags[nationalityID.String]; ok {
			props.FlagURL = flagURL
		}
	}
	return props
}

func studentSearchParams(q string) queries.SearchStudentsByNameParams {
	like := sql.NullString{String: q, Valid: true}
	return queries.SearchStudentsByNameParams{Column1: like, Column2: like}
}

func nationalityFieldsForStudent(ctx context.Context, nationalityID sql.NullString) (frontend.NationalityOption, bool) {
	if !nationalityID.Valid || nationalityID.String == "" {
		return frontend.NationalityOption{}, false
	}
	row, err := dbRO.GetQueries().GetNationalityByID(ctx, nationalityID.String)
	if err != nil {
		return frontend.NationalityOption{}, false
	}
	lang := ""
	if row.Language.Valid {
		lang = row.Language.String
	}
	return frontend.NationalityOption{
		ID:          row.ID,
		CountryName: row.CountryName,
		Language:    lang,
		FlagURL:     utils.URL(row.FlagPath),
	}, true
}
