package frontend

import "zion-english/internal/auth"

type NavItem struct {
	Path        string
	LinkID      string
	Title       string
	Description string
	FeatureCard bool
}

type navItemDef struct {
	Path         string
	LinkID       string
	Title        string
	TeacherTitle string
	Description  string
	TeacherDesc  string
	FeatureCard  bool
}

var navItemDefs = []navItemDef{
	{Path: "/profile", LinkID: "profileLink", Title: "My Profile", Description: "View account info and update settings"},
	{Path: "/teachers", LinkID: "teachersLink", Title: "Teachers", Description: "View and manage teachers"},
	{Path: "/students", LinkID: "studentsLink", Title: "Students", Description: "View and manage students"},
	{Path: "/classes", LinkID: "classesLink", Title: "Classes", TeacherTitle: "My Classes", Description: "View and record classes", TeacherDesc: "View and record your classes"},
	{Path: "/schedule", LinkID: "scheduleLink", Title: "Class Schedule", Description: "View and plan upcoming classes", FeatureCard: true},
	{Path: "/my-students", LinkID: "myStudentsLink", Title: "My Students", Description: "View your assigned students"},
	{Path: "/reports", LinkID: "reportsLink", Title: "Reports", Description: "View teacher payroll reports by cutoff period"},
	{Path: "/process", LinkID: "processLink", Title: "Process", Description: "Process CSV files and view logs"},
	{Path: "/logs", LinkID: "logsLink", Title: "Logs", TeacherTitle: "My Activity", Description: "View system logs", TeacherDesc: "View your recent actions"},
}

func NavItems(role auth.Role) []NavItem {
	var items []NavItem
	for _, def := range navItemDefs {
		if !IsNavAccessible(role, def.Path) {
			continue
		}
		title := def.Title
		if role == auth.RoleTeacher && def.TeacherTitle != "" {
			title = def.TeacherTitle
		}
		desc := def.Description
		if role == auth.RoleTeacher && def.TeacherDesc != "" {
			desc = def.TeacherDesc
		}
		items = append(items, NavItem{
			Path:        def.Path,
			LinkID:      def.LinkID,
			Title:       title,
			Description: desc,
			FeatureCard: def.FeatureCard,
		})
	}
	return items
}

func IsNavAccessible(role auth.Role, path string) bool {
	if role == "" {
		return false
	}

	if role == auth.RoleSuperuser {
		if path == "/my-students" {
			return false
		}
		return true
	}

	switch path {
	case "/students":
		return false
	case "/students/register":
		return true
	case "/classes", "/classes/record", "/schedule", "/profile", "/logs", "/my-students":
		return true
	default:
		return false
	}
}
