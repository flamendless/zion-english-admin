package cmd

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"unicode/utf8"

	"zion-english/frontend"
	"zion-english/internal/database/queries"
	"zion-english/internal/nationalities"
)

const studentLegalNameMaxRunes = 64

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
	iso := nationalities.NormalizeISO(raw)
	if iso == "" {
		return sql.NullString{}, nil
	}
	_, err := dbRO.GetQueries().GetNationalityByID(ctx, string(iso))
	if err != nil {
		return sql.NullString{}, ErrInvalidNationality
	}
	return sql.NullString{String: string(iso), Valid: true}, nil
}

func buildStudentAvatarProps(name, assignedColor string, isTestSeed bool, size string, nationalityID sql.NullString) frontend.AvatarProps {
	props := frontend.StudentAvatarProps(name, assignedColor, isTestSeed, size)
	if nationalityID.Valid && nationalityID.String != "" {
		props.NationalityISO = string(nationalities.NormalizeISO(nationalityID.String))
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
	}, true
}

func handleNationalityFlagPartial(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	iso := r.URL.Query().Get("nationalityId")
	writeHTML(w)
	if err := frontend.NationalityFlagPreview(iso).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}
