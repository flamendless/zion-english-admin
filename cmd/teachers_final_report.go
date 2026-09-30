package cmd

import (
	"database/sql"
	"errors"
	"net/http"
	"zion-english/internal/constants"
	"zion-english/internal/utils"
)

func handleTeacherFinalReport(w http.ResponseWriter, r *http.Request, teacherID int64) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	row, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, teacherID)
	if err != nil {
		HttpError(w, MsgTeacherNotFound, http.StatusNotFound)
		return
	}
	if row.Status != string(constants.TeacherStatusResigned) {
		HttpError(w, ErrTeacherNotResigned.Error(), http.StatusBadRequest)
		return
	}
	if !row.ResignedAt.Valid || row.ResignedAt.String == "" {
		HttpError(w, ErrTeacherResignationDateMissing.Error(), http.StatusBadRequest)
		return
	}

	startDate, endDate, err := utils.FinalReportRangeFromResignationDate(row.ResignedAt.String)
	if err != nil {
		HttpError(w, "Failed to determine final report period", http.StatusBadRequest)
		return
	}

	filename, _, _, err := generateTeacherReportFile(ctx, teacherID, startDate, endDate)
	if err != nil {
		if errors.Is(err, ErrReportNoRecords) {
			HttpError(w, MsgFinalReportNoRecords, http.StatusBadRequest)
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			HttpError(w, MsgTeacherNotFound, http.StatusNotFound)
			return
		}
		HttpError(w, "Failed to generate final report", http.StatusInternalServerError)
		return
	}

	auditReportDownload(ctx, filename)
	if !serveReportDownload(w, r, filename) {
		HttpError(w, "Failed to download final report", http.StatusInternalServerError)
	}
}
