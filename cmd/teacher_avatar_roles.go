package cmd

import (
	"context"

	"zion-english/frontend"
	"zion-english/internal/constants"
	"zion-english/internal/models"
	"zion-english/internal/teachers"
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

func avatarWithTeacherRoles(props frontend.AvatarProps, roles []constants.TeacherRole) frontend.AvatarProps {
	return frontend.WithRoleBadge(props, roles)
}

func avatarViewWithTeacherRoles(view models.AvatarView, roles []constants.TeacherRole) models.AvatarView {
	if role, ok := teachers.PrimaryTeacherRole(roles); ok {
		view.RoleBadge = string(role)
	}
	return view
}

func enrichClassRecordViewsWithRoleBadges(views []models.ClassRecordView, rolesMap map[int64][]constants.TeacherRole) {
	for i := range views {
		views[i].TeacherAvatar = avatarViewWithTeacherRoles(views[i].TeacherAvatar, rolesMap[views[i].TeacherID])
	}
}

func enrichScheduledClassViewsWithRoleBadges(views []models.ScheduledClassView, rolesMap map[int64][]constants.TeacherRole) {
	for i := range views {
		views[i].TeacherAvatar = avatarViewWithTeacherRoles(views[i].TeacherAvatar, rolesMap[views[i].TeacherID])
	}
}

func enrichDocumentItemsWithRoleBadges(items []frontend.DocumentItem, teacherIDs []int64, rolesMap map[int64][]constants.TeacherRole) {
	for i, id := range teacherIDs {
		items[i].UploadedByAvatar = avatarWithTeacherRoles(items[i].UploadedByAvatar, rolesMap[id])
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
