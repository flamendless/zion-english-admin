package frontend

import (
	"context"
	"strconv"

	"zion-english/internal/auth"
	"zion-english/internal/utils"
)

func ClassWizardPageDataForContext(ctx context.Context) ClassWizardPageData {
	return ClassWizardPageDataForUser(auth.GetRole(ctx), auth.GetUser(ctx))
}

func ClassWizardPageDataForUser(role auth.Role, user auth.User) ClassWizardPageData {
	data := ClassWizardPageData{
		TodayPHT: utils.TodayPHT(),
	}
	if auth.IsTeacherScoped(role) {
		data.LockTeacher = true
		data.TeacherID = strconv.FormatInt(user.ID, 10)
		data.TeacherName = user.Name
	} else if role != "" {
		data.ShowAllTeachers = true
	}
	return data
}
