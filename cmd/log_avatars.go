package cmd

import (
	"context"
	"database/sql"
	"zion-english/frontend"
	"zion-english/internal/database/queries"
)

func logCreatorAvatarCache(ctx context.Context, teacherIDs []int64) map[int64]frontend.AvatarProps {
	cache := make(map[int64]frontend.AvatarProps, len(teacherIDs))
	for _, id := range teacherIDs {
		if id <= 0 {
			continue
		}
		if _, loaded := cache[id]; loaded {
			continue
		}
		t, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, id)
		if err != nil {
			continue
		}
		cache[id] = buildTeacherListAvatarProps(
			t.ID,
			t.FirstName,
			t.MiddleName,
			t.LastName,
			t.AssignedColor,
			t.ProfilePicture,
		)
	}
	return cache
}

func uniqueTeacherIDsFromNullInt64(ids []sql.NullInt64) []int64 {
	seen := make(map[int64]struct{})
	out := make([]int64, 0)
	for _, id := range ids {
		if !id.Valid || id.Int64 <= 0 {
			continue
		}
		if _, ok := seen[id.Int64]; ok {
			continue
		}
		seen[id.Int64] = struct{}{}
		out = append(out, id.Int64)
	}
	return out
}

func avatarForLogCreator(createdBy sql.NullInt64, displayName string, cache map[int64]frontend.AvatarProps) frontend.AvatarProps {
	if createdBy.Valid && createdBy.Int64 > 0 {
		if props, ok := cache[createdBy.Int64]; ok {
			return props
		}
	}
	return notificationNameAvatar(displayName)
}

func systemLogCreatorIDsFromAllRows(rows []queries.GetAllLogsFilteredRow) []int64 {
	ids := make([]sql.NullInt64, len(rows))
	for i, row := range rows {
		ids[i] = row.CreatedBy
	}
	return uniqueTeacherIDsFromNullInt64(ids)
}

func systemLogCreatorIDsFromTeacherRows(rows []queries.GetLogsByCreatedByFilteredRow) []int64 {
	ids := make([]sql.NullInt64, len(rows))
	for i, row := range rows {
		ids[i] = row.CreatedBy
	}
	return uniqueTeacherIDsFromNullInt64(ids)
}

func uploadLogCreatorIDsFromAllRows(rows []queries.GetUploadLogsFilteredRow) []int64 {
	ids := make([]sql.NullInt64, len(rows))
	for i, row := range rows {
		ids[i] = row.CreatedBy
	}
	return uniqueTeacherIDsFromNullInt64(ids)
}

func uploadLogCreatorIDsFromTeacherRows(rows []queries.GetUploadLogsByCreatedByFilteredRow) []int64 {
	ids := make([]sql.NullInt64, len(rows))
	for i, row := range rows {
		ids[i] = row.CreatedBy
	}
	return uniqueTeacherIDsFromNullInt64(ids)
}
