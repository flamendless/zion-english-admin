package cmd

import (
	"database/sql"
	"regexp"
	"strconv"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/utils"
)

var uploadLogFilenameFromSummaryRE = regexp.MustCompile(`file '([^']*)'`)

func mapUploadLogItemFromFiltered(row queries.GetUploadLogsFilteredRow) frontend.UploadLogItem {
	return mapUploadLogFields(
		row.ID,
		row.Module,
		row.Outcome,
		row.Kind,
		row.Summary,
		row.Filename,
		row.FileSize,
		row.CompressPreset,
		row.CreatedByName,
		row.CreatedAt,
	)
}

func mapUploadLogItemFromTeacherFiltered(row queries.GetUploadLogsByCreatedByFilteredRow) frontend.UploadLogItem {
	return mapUploadLogFields(
		row.ID,
		row.Module,
		row.Outcome,
		row.Kind,
		row.Summary,
		row.Filename,
		row.FileSize,
		row.CompressPreset,
		row.CreatedByName,
		row.CreatedAt,
	)
}

func mapUploadLogFields(
	id int64,
	module string,
	outcome string,
	kind string,
	summary string,
	filename sql.NullString,
	fileSize sql.NullInt64,
	compressPreset sql.NullString,
	createdBy string,
	createdAt string,
) frontend.UploadLogItem {
	kindEnum := constants.UploadLogKind(kind)
	if !constants.ValidUploadLogKind(kind) {
		kindEnum = inferUploadLogKindFromSummary(summary)
	}

	filenameDisplay := "-"
	if filename.Valid && filename.String != "" {
		filenameDisplay = filename.String
	} else {
		parsed := parseUploadLogFilenameFromSummary(summary)
		if parsed != "" {
			filenameDisplay = parsed
		}
	}

	outcomeEnum := constants.UploadLogOutcome(outcome)
	if outcomeEnum != constants.UploadLogOutcomeSucceeded && outcomeEnum != constants.UploadLogOutcomeFailed {
		outcomeEnum = constants.UploadLogOutcomeFailed
	}

	return frontend.UploadLogItem{
		ID:              strconv.FormatInt(id, 10),
		Module:          module,
		Outcome:         outcomeEnum,
		Kind:            kindEnum,
		Summary:         summary,
		Filename:        filenameDisplay,
		FileSizeDisplay: formatUploadLogFileSize(fileSize),
		PresetDisplay:   formatUploadLogCompressPreset(compressPreset),
		CreatedBy:       createdBy,
		CreatedAt:       createdAt,
	}
}

func inferUploadLogKindFromSummary(summary string) constants.UploadLogKind {
	lower := strings.ToLower(summary)
	switch {
	case strings.Contains(lower, "intro video"):
		return constants.UploadLogKindIntroVideo
	case strings.Contains(lower, "profile picture"):
		return constants.UploadLogKindAvatar
	case strings.Contains(lower, "id document"), strings.Contains(lower, "resume/cv"):
		return constants.UploadLogKindDocument
	default:
		return constants.UploadLogKindOther
	}
}

func parseUploadLogFilenameFromSummary(summary string) string {
	match := uploadLogFilenameFromSummaryRE.FindStringSubmatch(summary)
	if len(match) < 2 {
		return ""
	}
	return match[1]
}

func formatUploadLogFileSize(size sql.NullInt64) string {
	if !size.Valid {
		return "-"
	}
	return utils.FormatFileSize(size.Int64)
}

func formatUploadLogCompressPreset(preset sql.NullString) string {
	if !preset.Valid || preset.String == "" {
		return "-"
	}
	if !constants.ValidIntroVideoCompressPreset(preset.String) {
		return preset.String
	}
	return constants.IntroVideoCompressPresetLabel(constants.IntroVideoCompressPreset(preset.String))
}
