package notifications

import (
	"context"
	"fmt"
	"time"
	"zion-english/internal/utils"
)

func (s *Service) ScanMissedClasses(ctx context.Context) {
	nowPHT := utils.DateTimePHT(time.Now())
	rows, err := s.q.GetMissedScheduledClasses(ctx, nowPHT)
	if err != nil {
		return
	}
	system := SystemUser()
	for _, row := range rows {
		startTime := "00:00"
		if row.StartTime.Valid && row.StartTime.String != "" {
			startTime = row.StartTime.String
		}
		message := fmt.Sprintf(
			"Missed scheduled class with %s on %s at %s",
			row.StudentName,
			row.ScheduledDate,
			startTime,
		)
		dedupe := fmt.Sprintf("missed_class:%d", row.ID)
		s.NotifyTeacher(ctx, row.TeacherID, row.TeacherName, system, KindMissedClass, message, dedupe)
		s.NotifySuperuser(ctx, system, KindMissedClass,
			fmt.Sprintf("Teacher %s missed class with %s on %s at %s",
				row.TeacherName, row.StudentName, row.ScheduledDate, startTime),
			dedupe+":superuser",
		)
	}
}
