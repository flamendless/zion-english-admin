package notifications

import (
	"context"
	"zion-english/internal/auth"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"

	"go.uber.org/zap"
)

const (
	superuserName = "superuser"
	panelLimit    = 10
)

type Service struct {
	q *queries.Queries
}

func New(q *queries.Queries) *Service {
	return &Service{q: q}
}

func (s *Service) NotifySuperuser(ctx context.Context, from auth.User, kind, message, dedupeKey string) {
	s.insert(ctx, from, nil, superuserName, kind, message, dedupeKey)
}

func (s *Service) NotifyTeacher(ctx context.Context, teacherID int64, teacherName string, from auth.User, kind, message, dedupeKey string) {
	if teacherID <= 0 {
		return
	}
	s.insert(ctx, from, teacherID, teacherName, kind, message, dedupeKey)
}

func (s *Service) NotifyTeachers(ctx context.Context, teacherIDs []int64, teacherNames map[int64]string, from auth.User, kind, message string) {
	for _, id := range teacherIDs {
		name := teacherNames[id]
		if name == "" {
			name = "teacher"
		}
		s.NotifyTeacher(ctx, id, name, from, kind, message, "")
	}
}

func (s *Service) insert(ctx context.Context, from auth.User, toTeacherID interface{}, toName, kind, message, dedupeKey string) {
	fromTeacherID, fromName := fromParams(from)
	var dedupe interface{}
	if dedupeKey != "" {
		dedupe = dedupeKey
	}
	if err := s.q.InsertNotification(ctx, queries.InsertNotificationParams{
		FromTeacherID: fromTeacherID,
		FromName:      fromName,
		ToTeacherID:   toTeacherID,
		ToName:        toName,
		Message:       message,
		Kind:          kind,
		DedupeKey:     dedupe,
	}); err != nil {
		logs.Log().Info("notifications", zap.Error(err), zap.String("kind", kind))
	}
}

func fromParams(from auth.User) (interface{}, string) {
	name := from.Name
	if name == "" {
		name = "system"
	}
	if from.ID > 0 {
		return from.ID, name
	}
	return nil, name
}

func SystemUser() auth.User {
	return auth.User{Name: "System"}
}

func (s *Service) UnreadCount(ctx context.Context, user auth.User) (int64, error) {
	if user.Role == auth.RoleSuperuser {
		return s.q.CountUnreadNotificationsForSuperuser(ctx)
	}
	return s.q.CountUnreadNotificationsForTeacher(ctx, user.ID)
}

func (s *Service) Recent(ctx context.Context, user auth.User) ([]queries.TblNotification, error) {
	if user.Role == auth.RoleSuperuser {
		return s.q.GetRecentNotificationsForSuperuser(ctx, panelLimit)
	}
	return s.q.GetRecentNotificationsForTeacher(ctx, queries.GetRecentNotificationsForTeacherParams{
		ToTeacherID: user.ID,
		Limit:       panelLimit,
	})
}

func (s *Service) ListPaged(ctx context.Context, user auth.User, unreadOnly bool, limit, offset int64) ([]queries.TblNotification, error) {
	filter := unreadFilter(unreadOnly)
	if user.Role == auth.RoleSuperuser {
		return s.q.GetNotificationsPagedForSuperuser(ctx, queries.GetNotificationsPagedForSuperuserParams{
			Column1: filter,
			Limit:   limit,
			Offset:  offset,
		})
	}
	return s.q.GetNotificationsPagedForTeacher(ctx, queries.GetNotificationsPagedForTeacherParams{
		ToTeacherID: user.ID,
		Column2:     filter,
		Limit:       limit,
		Offset:      offset,
	})
}

func (s *Service) Count(ctx context.Context, user auth.User, unreadOnly bool) (int64, error) {
	filter := unreadFilter(unreadOnly)
	if user.Role == auth.RoleSuperuser {
		return s.q.CountNotificationsForSuperuser(ctx, filter)
	}
	return s.q.CountNotificationsForTeacher(ctx, queries.CountNotificationsForTeacherParams{
		ToTeacherID: user.ID,
		Column2:     filter,
	})
}

func (s *Service) MarkRead(ctx context.Context, user auth.User, id int64) error {
	if user.Role == auth.RoleSuperuser {
		return s.q.MarkNotificationReadForSuperuser(ctx, id)
	}
	return s.q.MarkNotificationReadForTeacher(ctx, queries.MarkNotificationReadForTeacherParams{
		ID:          id,
		ToTeacherID: user.ID,
	})
}

func (s *Service) MarkAllRead(ctx context.Context, user auth.User) error {
	if user.Role == auth.RoleSuperuser {
		return s.q.MarkAllNotificationsReadForSuperuser(ctx)
	}
	return s.q.MarkAllNotificationsReadForTeacher(ctx, user.ID)
}

func unreadFilter(unreadOnly bool) interface{} {
	if unreadOnly {
		return 1
	}
	return 0
}
