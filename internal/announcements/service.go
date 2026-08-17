package announcements

import (
	"context"

	"zion-english/internal/database/queries"
)

func ReplaceTeacherLinks(ctx context.Context, db *queries.Queries, announcementID int64, teacherIDs []int64) error {
	if err := db.DeleteAnnouncementTeacherLinks(ctx, announcementID); err != nil {
		return err
	}
	for _, teacherID := range teacherIDs {
		if err := db.InsertAnnouncementTeacherM2M(ctx, queries.InsertAnnouncementTeacherM2MParams{
			AnnouncementID: announcementID,
			TeacherID:      teacherID,
		}); err != nil {
			return err
		}
	}
	return nil
}
