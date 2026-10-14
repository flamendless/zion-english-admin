package frontend

import (
	"fmt"

	"zion-english/internal/utils"
)

func HubStudentsFullPageURL(viewAsAdmin bool, teacherID string) string {
	if viewAsAdmin {
		return fmt.Sprintf("%s?teacherId=%s", utils.URL("/students"), teacherID)
	}
	return utils.URL("/my-students")
}
