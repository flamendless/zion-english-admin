package cmd

import (
	"context"

	"zion-english/frontend"
	"zion-english/internal/constants"
	"zion-english/internal/models"
)

func loadRolesByTeacherIDs(ctx context.Context, teacherIDs []int64) (map[int64][]constants.TeacherRole, error) {
	rolesByTeacher := make(map[int64][]constants.TeacherRole)
	if len(teacherIDs) == 0 {
		return rolesByTeacher, nil
	}

	roleRows, err := dbRO.GetQueries().GetTeacherRolesByTeacherIDs(ctx, teacherIDs)
	if err != nil {
		return nil, err
	}
	for _, row := range roleRows {
		rolesByTeacher[row.TeacherID] = append(rolesByTeacher[row.TeacherID], constants.TeacherRole(row.Role))
	}
	return rolesByTeacher, nil
}

func loadTeacherStatusesByIDs(ctx context.Context, teacherIDs []int64) (map[int64]constants.TeacherStatus, error) {
	statusByTeacher := make(map[int64]constants.TeacherStatus)
	if len(teacherIDs) == 0 {
		return statusByTeacher, nil
	}

	rows, err := dbRO.GetQueries().GetTeacherStatusesByIDs(ctx, teacherIDs)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		statusByTeacher[row.ID] = constants.TeacherStatus(row.Status)
	}
	return statusByTeacher, nil
}

func teacherStatusFromMap(statusMap map[int64]constants.TeacherStatus, teacherID int64) constants.TeacherStatus {
	if status, ok := statusMap[teacherID]; ok {
		return status
	}
	return constants.TeacherStatusApproved
}

func teacherStatusForID(ctx context.Context, teacherID int64) (constants.TeacherStatus, error) {
	statusMap, err := loadTeacherStatusesByIDs(ctx, []int64{teacherID})
	if err != nil {
		return "", err
	}
	return teacherStatusFromMap(statusMap, teacherID), nil
}

func avatarWithTeacherRoles(props frontend.AvatarProps, roles []constants.TeacherRole, status constants.TeacherStatus) frontend.AvatarProps {
	return frontend.AssignedTeacherAvatar(props, roles, status)
}

func avatarViewWithTeacherRoles(view models.AvatarView, roles []constants.TeacherRole, status constants.TeacherStatus) models.AvatarView {
	return frontend.ApplyTeacherAvatarViewRoles(view, roles, status)
}

func enrichClassRecordViewsWithRoleBadges(views []models.ClassRecordView, rolesMap map[int64][]constants.TeacherRole, statusMap map[int64]constants.TeacherStatus) {
	for i := range views {
		views[i].TeacherAvatar = avatarViewWithTeacherRoles(views[i].TeacherAvatar, rolesMap[views[i].TeacherID], teacherStatusFromMap(statusMap, views[i].TeacherID))
	}
}

func enrichScheduledClassViewsWithRoleBadges(views []models.ScheduledClassView, rolesMap map[int64][]constants.TeacherRole, statusMap map[int64]constants.TeacherStatus) {
	for i := range views {
		views[i].TeacherAvatar = avatarViewWithTeacherRoles(views[i].TeacherAvatar, rolesMap[views[i].TeacherID], teacherStatusFromMap(statusMap, views[i].TeacherID))
	}
}

func enrichDocumentItemsWithRoleBadges(items []frontend.DocumentItem, teacherIDs []int64, rolesMap map[int64][]constants.TeacherRole, statusMap map[int64]constants.TeacherStatus) {
	for i, id := range teacherIDs {
		items[i].UploadedByAvatar = avatarWithTeacherRoles(items[i].UploadedByAvatar, rolesMap[id], teacherStatusFromMap(statusMap, id))
	}
}

func enrichIntroVideoItemsWithRoleBadges(items []frontend.IntroVideoItem, teacherIDs []int64, rolesMap map[int64][]constants.TeacherRole, statusMap map[int64]constants.TeacherStatus) {
	for i, id := range teacherIDs {
		items[i].UploadedByAvatar = avatarWithTeacherRoles(items[i].UploadedByAvatar, rolesMap[id], teacherStatusFromMap(statusMap, id))
	}
}

func uniqueTeacherIDs(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	unique := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return unique
}
