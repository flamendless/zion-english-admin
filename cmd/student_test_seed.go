package cmd

import (
	"context"
	"database/sql"
	"strconv"

	"zion-english/internal/models"
)

func studentTestSeedFilterTriplet(filter string) (string, string, string) {
	return filter, filter, filter
}

func studentAvatarView(name, assignedColor string, isTestSeed int64, nationalityID sql.NullString) models.AvatarView {
	props := buildStudentAvatarProps(name, assignedColor, isTestSeed == 1, "sm", nationalityID)
	return models.AvatarView{
		Initials:       props.Initials,
		AssignedColor:  props.AssignedColor,
		PictureURL:     props.PictureURL,
		HasPicture:     props.HasPicture,
		Alt:            props.Alt,
		RoleBadge:      props.RoleBadge,
		NationalityISO: props.NationalityISO,
	}
}

func parseIsTestSeedForm(value string) int64 {
	if value == "1" {
		return 1
	}
	return 0
}

func appendSelectedStudentIfMissing(ctx context.Context, responses []models.StudentAPIResponse, selectedID string) []models.StudentAPIResponse {
	if selectedID == "" {
		return responses
	}
	id, err := strconv.ParseInt(selectedID, 10, 64)
	if err != nil || id <= 0 {
		return responses
	}
	for _, s := range responses {
		if s.ID == id {
			return responses
		}
	}
	row, err := dbRO.GetQueries().GetStudentByID(ctx, id)
	if err != nil {
		return responses
	}
	return append(responses, models.StudentAPIResponse{
		ID:           row.ID,
		Name:         studentDisplayName(row.Name, row.StudentNumber, row.LegalName),
		Currency:     row.Currency,
		RatePerClass: row.RatePerClass,
	})
}
