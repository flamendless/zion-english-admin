package cmd

import (
	"context"
	"zion-english/internal/auth"
	"zion-english/internal/notifications"
	"zion-english/internal/utils"
)

var notifySvc *notifications.Service

func initNotifyService() {
	notifySvc = notifications.New(dbRW.GetQueries())
}

func notifySuperuser(ctx context.Context, from auth.User, kind, message, dedupeKey string) {
	if notifySvc == nil {
		return
	}
	notifySvc.NotifySuperuser(ctx, from, kind, message, dedupeKey)
}

func notifyTeacher(ctx context.Context, teacherID int64, teacherName string, from auth.User, kind, message, dedupeKey string) {
	if notifySvc == nil {
		return
	}
	notifySvc.NotifyTeacher(ctx, teacherID, teacherName, from, kind, message, dedupeKey)
}

func notifyTeachers(ctx context.Context, teacherIDs []int64, teacherNames map[int64]string, from auth.User, kind, message string) {
	if notifySvc == nil {
		return
	}
	notifySvc.NotifyTeachers(ctx, teacherIDs, teacherNames, from, kind, message)
}

func notifyCrossParty(ctx context.Context, actor auth.User, teacherID int64, teacherName string, kind, message string) {
	if actor.Role == auth.RoleTeacher {
		notifySuperuser(ctx, actor, kind, message, "")
		return
	}
	if actor.Role == auth.RoleSuperuser && teacherID > 0 {
		notifyTeacher(ctx, teacherID, teacherName, actor, kind, message, "")
	}
}

func teacherNameByID(ctx context.Context, teacherID int64) string {
	t, err := dbRO.GetQueries().GetTeacherFullByID(ctx, teacherID)
	if err != nil {
		return "teacher"
	}
	return utils.ComposePersonName(t.FirstName, t.MiddleName, t.LastName)
}

func teacherNamesMap(ctx context.Context, ids []int64) map[int64]string {
	names := make(map[int64]string, len(ids))
	for _, id := range ids {
		names[id] = teacherNameByID(ctx, id)
	}
	return names
}

func notifyNewlyAssignedTeachers(ctx context.Context, actor auth.User, studentName string, before []int64, after []int64) {
	beforeSet := make(map[int64]struct{}, len(before))
	for _, id := range before {
		beforeSet[id] = struct{}{}
	}
	var added []int64
	for _, id := range after {
		if _, ok := beforeSet[id]; !ok {
			added = append(added, id)
		}
	}
	if len(added) == 0 {
		return
	}
	notifyTeachers(ctx, added, teacherNamesMap(ctx, added), actor, notifications.KindStudentUpdated,
		"Student '"+studentName+"' was assigned to you")
}
