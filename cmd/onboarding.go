package cmd

import (
	"context"
	"database/sql"
	"errors"

	"zion-english/internal/constants"
	"zion-english/internal/featureflags"
	"zion-english/internal/onboarding"
)

func loadTeacherOnboardingChecklist(ctx context.Context, teacherID int64) ([]onboarding.Item, error) {
	row, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, teacherID)
	if err != nil {
		return nil, err
	}

	docsStatus := ""
	docRows, err := dbRO.GetQueries().GetLatestTeacherDocumentStatusesByTeacherIDs(ctx, []int64{teacherID})
	if err != nil {
		return nil, err
	}
	for _, docRow := range docRows {
		if docRow.TeacherID == teacherID {
			docsStatus = docRow.Status
			break
		}
	}

	introStatus := ""
	roles, err := loadTeacherRoles(ctx, teacherID)
	if err != nil {
		return nil, err
	}
	introVisible, introUploadAllowed := featureflags.IntroVideoUploadAccess(ctx, dbRO, roles)
	if introVisible {
		introRow, err := dbRO.GetQueries().GetLatestTeacherIntroVideoByTeacherID(ctx, teacherID)
		if err == nil {
			introStatus = introRow.Status
		} else if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}
	introRequired := introVisible && (introUploadAllowed || introStatus != "")

	trainingTotal, err := dbRO.GetQueries().CountTrainingMaterialsRequiredPublished(ctx)
	if err != nil {
		return nil, err
	}
	trainingCompleted, err := dbRO.GetQueries().CountRequiredPublishedTrainingMaterialsCompletedByTeacher(ctx, teacherID)
	if err != nil {
		return nil, err
	}

	zoomConnected, zoomConfigured, _, _ := profileZoomStatus(ctx, teacherID)
	googleConnected, googleConfigured, _, _ := profileGoogleCalendarStatus(ctx, teacherID)
	zoomVisibleToTeacher := featureflags.IsVisibleToTeacherRoles(ctx, dbRO, constants.FeatureFlagIntegrationZoom, roles)
	googleVisibleToTeacher := featureflags.IsVisibleToTeacherRoles(ctx, dbRO, constants.FeatureFlagIntegrationGoogleCalendar, roles)

	return onboarding.Build(onboarding.Input{
		TeacherStatus:          constants.TeacherStatus(row.Status),
		HasProfilePhoto:        row.ProfilePicture.Valid && row.ProfilePicture.String != "",
		DocsStatus:             docsStatus,
		IntroVideoStatus:       introStatus,
		IntroVideoRequired:     introRequired,
		TrainingRequiredCompleted: trainingCompleted,
		TrainingRequiredTotal:   trainingTotal,
		ZoomShow:               zoomConfigured && (zoomConnected || zoomVisibleToTeacher),
		ZoomConnected:          zoomConnected,
		GoogleShow:             googleConfigured && (googleConnected || googleVisibleToTeacher),
		GoogleConnected:        googleConnected,
	}), nil
}
