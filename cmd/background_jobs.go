package cmd

import (
	"context"
	"net/http"
	"strconv"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/utils"
)

func backgroundJobsForTeacher(ctx context.Context, teacherID int64) ([]frontend.PersistentJobView, error) {
	rows, err := dbRO.GetQueries().GetProcessingTeacherIntroVideosByTeacherID(ctx, teacherID)
	if err != nil {
		return nil, err
	}
	jobs := make([]frontend.PersistentJobView, 0, len(rows))
	for _, row := range rows {
		detail := "-"
		if row.OriginalFilename.Valid && row.OriginalFilename.String != "" {
			detail = row.OriginalFilename.String
		}
		jobs = append(jobs, frontend.PersistentJobView{
			ID:            "intro-video-" + strconv.FormatInt(row.ID, 10),
			Label:         "Intro video",
			Detail:        detail,
			Phase:         frontend.PersistentJobPhaseProcessing,
			Indeterminate: true,
		})
	}
	return jobs, nil
}

func handlePersistentBackgroundJobsPartial(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	if auth.GetRole(ctx) != auth.RoleTeacher {
		return
	}

	user := auth.GetUser(ctx)
	jobs, err := backgroundJobsForTeacher(ctx, user.ID)
	if err != nil {
		return
	}
	if len(jobs) == 0 {
		return
	}

	writeHTML(w)
	panel := frontend.BuildPersistentJobPanel(jobs)
	if err := frontend.PersistentBackgroundJobsPoll(utils.URL("/background-jobs/persistent"), panel).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}
