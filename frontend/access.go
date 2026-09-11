package frontend

import "zion-english/internal/auth"

type NavItem struct {
	Path          string
	LinkID        string
	Title         string
	Description   string
	Icon          NavIconKind
	FeatureCard   bool
	AdminOnlyCard bool
}

type NavBarEntryKind string

const (
	NavBarEntryLink  NavBarEntryKind = "link"
	NavBarEntryGroup NavBarEntryKind = "group"
)

type NavBarEntry struct {
	Kind  NavBarEntryKind
	Item  NavItem
	Group NavGroup
}

type NavGroup struct {
	ID            string
	Label         string
	Description   string
	Icon          NavIconKind
	FeatureCard   bool
	AdminOnlyCard bool
	Items         []NavItem
}

type navItemDef struct {
	Path          string
	LinkID        string
	Title         string
	TeacherTitle  string
	Description   string
	TeacherDesc   string
	FeatureCard   bool
	AdminOnlyCard bool
	HideFromNav   bool
}

type navGroupDef struct {
	ID            string
	Label         string
	Description   string
	TeacherDesc   string
	FeatureCard   bool
	AdminOnlyCard bool
	Paths         []string
}

type navLayoutEntry struct {
	LinkPath string
	GroupID  string
}

var dashboardNavItem = NavItem{
	Path:        "/dashboard",
	LinkID:      "dashboardLink",
	Title:       "Dashboard",
	Description: "Overview and quick actions",
	Icon:        NavIconDefault,
}

var navItemDefs = []navItemDef{
	{Path: "/profile", LinkID: "profileLink", Title: "My Profile", Description: "View account info and update settings"},
	{Path: "/learning-materials", LinkID: "learningMaterialsLink", Title: "Learning Library", Description: "Browse and share teaching materials and resources", FeatureCard: true},
	{Path: "/documents", LinkID: "documentsLink", Title: "Documents", TeacherTitle: "My Documents", Description: "Review teacher uploads and ID documents", TeacherDesc: "View your uploaded profile photos and ID documents"},
	{Path: "/guides", LinkID: "guidesLink", Title: "Guides", Description: "Step-by-step help for using the admin tool"},
	{Path: "/teachers", LinkID: "teachersLink", Title: "Teachers", Description: "View and manage teachers", AdminOnlyCard: true},
	{Path: "/students", LinkID: "studentsLink", Title: "Students", Description: "View and manage students", AdminOnlyCard: true},
	{Path: "/classes", LinkID: "classesLink", Title: "Classes", TeacherTitle: "My Classes", Description: "View and record classes", TeacherDesc: "View and record your classes", HideFromNav: true},
	{Path: "/schedule", LinkID: "scheduleLink", Title: "Class Schedule", Description: "View and plan upcoming classes", FeatureCard: true, HideFromNav: true},
	{Path: "/schedule/series", LinkID: "scheduleSeriesLink", Title: "Repeating Scheduled Classes", Description: "View and manage linked class series", FeatureCard: true, HideFromNav: true},
	{Path: "/my-students", LinkID: "myStudentsLink", Title: "My Students", Description: "View your assigned students"},
	{Path: "/reports", LinkID: "reportsLink", Title: "Reports", Description: "View teacher payroll reports by cutoff period", AdminOnlyCard: true},
	{Path: "/analytics", LinkID: "analyticsLink", Title: "Analytics", TeacherTitle: "My Analytics", Description: "Attendance, utilization, and student retention insights", TeacherDesc: "View attendance and utilization for your classes"},
	{Path: "/process", LinkID: "processLink", Title: "Process", Description: "Process CSV files and view logs", AdminOnlyCard: true},
	{Path: "/feature-flags", LinkID: "featureFlagsLink", Title: "Feature Flags", Description: "Toggle integration connection availability", AdminOnlyCard: true},
	{Path: "/logs", LinkID: "logsLink", Title: "Logs", TeacherTitle: "My Activity", Description: "View system logs", TeacherDesc: "View your recent actions"},
}

var navGroupDefs = []navGroupDef{
	{ID: "classes", Label: "Classes", Description: "View classes, schedules, and repeating series", TeacherDesc: "View and record classes, schedules, and repeating series", FeatureCard: true, Paths: []string{"/classes", "/schedule", "/schedule/series"}},
	{ID: "people", Label: "People", Description: "Manage teachers and students", AdminOnlyCard: true, Paths: []string{"/teachers", "/students"}},
	{ID: "resources", Label: "Resources", Description: "Guides, documents, and learning materials", TeacherDesc: "Guides, your documents, and learning materials", FeatureCard: true, Paths: []string{"/guides", "/documents", "/learning-materials"}},
	{ID: "insights", Label: "Insights", Description: "Payroll reports and analytics", AdminOnlyCard: true, Paths: []string{"/reports", "/analytics"}},
	{ID: "admin", Label: "Admin", Description: "Process CSV files, feature flags, and system logs", AdminOnlyCard: true, Paths: []string{"/process", "/feature-flags", "/logs"}},
}

var adminNavLayout = []navLayoutEntry{
	{LinkPath: "/dashboard"},
	{LinkPath: "/profile"},
	{GroupID: "resources"},
	{GroupID: "people"},
	{GroupID: "classes"},
	{GroupID: "insights"},
	{GroupID: "admin"},
}

var teacherNavLayout = []navLayoutEntry{
	{LinkPath: "/dashboard"},
	{LinkPath: "/profile"},
	{GroupID: "resources"},
	{LinkPath: "/my-students"},
	{GroupID: "classes"},
	{LinkPath: "/analytics"},
	{LinkPath: "/logs"},
}

var testerNavLayout = []navLayoutEntry{
	{LinkPath: "/dashboard"},
	{LinkPath: "/profile"},
	{GroupID: "classes"},
}

var groupedNavPaths = buildGroupedNavPaths()

func NavItems(role auth.Role) []NavItem {
	return navItemsForRole(role, false)
}

func NavBarItems(role auth.Role) []NavItem {
	return navItemsForRole(role, true)
}

func NavBarEntries(role auth.Role) []NavBarEntry {
	return navEntriesFromLayout(role, false)
}

func DashboardQuickActionEntries(role auth.Role) []NavBarEntry {
	return navEntriesFromLayout(role, true)
}

func navEntriesFromLayout(role auth.Role, skipDashboard bool) []NavBarEntry {
	layout := navLayoutForRole(role)
	var entries []NavBarEntry
	for _, layoutEntry := range layout {
		if layoutEntry.LinkPath != "" {
			if skipDashboard && layoutEntry.LinkPath == "/dashboard" {
				continue
			}
			item := navItemForPath(role, layoutEntry.LinkPath)
			if item == nil {
				continue
			}
			entries = append(entries, NavBarEntry{
				Kind: NavBarEntryLink,
				Item: *item,
			})
			continue
		}
		group := navGroupForRole(role, layoutEntry.GroupID)
		if group == nil || len(group.Items) < 2 {
			continue
		}
		entries = append(entries, NavBarEntry{
			Kind:  NavBarEntryGroup,
			Group: *group,
		})
	}
	return entries
}

func NavGroups(role auth.Role) []NavGroup {
	layout := navLayoutForRole(role)
	seen := make(map[string]bool)
	var groups []NavGroup
	for _, layoutEntry := range layout {
		if layoutEntry.GroupID == "" || seen[layoutEntry.GroupID] {
			continue
		}
		group := navGroupForRole(role, layoutEntry.GroupID)
		if group == nil || len(group.Items) < 2 {
			continue
		}
		seen[layoutEntry.GroupID] = true
		groups = append(groups, *group)
	}
	return groups
}

func DashboardStandaloneItems(role auth.Role) []NavItem {
	var items []NavItem
	for _, def := range navItemDefs {
		if groupedNavPaths[def.Path] {
			continue
		}
		if !IsNavAccessible(role, def.Path) {
			continue
		}
		items = append(items, navItemFromDef(role, def))
	}
	return items
}

func DashboardExtraItems(role auth.Role) []NavItem {
	var items []NavItem
	if auth.HasAdminAccess(role) {
		items = append(items, NavItem{
			Path:          "/announcements",
			LinkID:        "announcementsLink",
			Title:         "Announcements",
			Description:   "Create and manage system-wide banners",
			Icon:          NavIconForPath("/announcements"),
			AdminOnlyCard: true,
		})
	}
	items = append(items, NavItem{
		Path:        "/changelogs",
		LinkID:      "changelogsLink",
		Title:       "Changelogs",
		Description: "See what's new in the admin tool",
		Icon:        NavIconForPath("/changelogs"),
	})
	return items
}

func navItemsForRole(role auth.Role, navbarOnly bool) []NavItem {
	var items []NavItem
	for _, def := range navItemDefs {
		if navbarOnly && def.HideFromNav {
			continue
		}
		if !IsNavAccessible(role, def.Path) {
			continue
		}
		items = append(items, navItemFromDef(role, def))
	}
	return items
}

func navLayoutForRole(role auth.Role) []navLayoutEntry {
	if auth.HasAdminAccess(role) {
		return adminNavLayout
	}
	if role == auth.RoleTester {
		return testerNavLayout
	}
	return teacherNavLayout
}

func navGroupForRole(role auth.Role, groupID string) *NavGroup {
	for _, def := range navGroupDefs {
		if def.ID != groupID {
			continue
		}
		var items []NavItem
		for _, path := range def.Paths {
			item := navItemForPath(role, path)
			if item == nil {
				continue
			}
			items = append(items, *item)
		}
		if len(items) == 0 {
			return nil
		}
		desc := def.Description
		if auth.IsTeacherScoped(role) && def.TeacherDesc != "" {
			desc = def.TeacherDesc
		}
		return &NavGroup{
			ID:            def.ID,
			Label:         def.Label,
			Description:   desc,
			Icon:          NavIconForGroupID(def.ID),
			FeatureCard:   def.FeatureCard,
			AdminOnlyCard: def.AdminOnlyCard,
			Items:         items,
		}
	}
	return nil
}

func navItemForPath(role auth.Role, path string) *NavItem {
	if path == "/dashboard" {
		if !IsNavAccessible(role, "/dashboard") {
			return nil
		}
		return &dashboardNavItem
	}
	for _, def := range navItemDefs {
		if def.Path != path {
			continue
		}
		if !IsNavAccessible(role, def.Path) {
			return nil
		}
		item := navItemFromDef(role, def)
		return &item
	}
	return nil
}

func navItemFromDef(role auth.Role, def navItemDef) NavItem {
	title := def.Title
	if auth.IsTeacherScoped(role) && def.TeacherTitle != "" {
		title = def.TeacherTitle
	}
	desc := def.Description
	if auth.IsTeacherScoped(role) && def.TeacherDesc != "" {
		desc = def.TeacherDesc
	}
	return NavItem{
		Path:          def.Path,
		LinkID:        def.LinkID,
		Title:         title,
		Description:   desc,
		Icon:          NavIconForPath(def.Path),
		FeatureCard:   def.FeatureCard,
		AdminOnlyCard: def.AdminOnlyCard,
	}
}

func buildGroupedNavPaths() map[string]bool {
	grouped := make(map[string]bool)
	for _, group := range navGroupDefs {
		for _, path := range group.Paths {
			grouped[path] = true
		}
	}
	return grouped
}

func IsNavAccessible(role auth.Role, path string) bool {
	if role == "" {
		return false
	}

	if path == "/dashboard" {
		return true
	}

	if auth.HasAdminAccess(role) {
		if path == "/my-students" {
			return false
		}
		if path == "/feature-flags" {
			return role == auth.RoleSuperuser
		}
		return true
	}

	if role == auth.RoleTester {
		switch path {
		case "/profile", "/classes", "/classes/record", "/schedule", "/schedule/record", "/schedule/repeat", "/schedule/series":
			return true
		default:
			return false
		}
	}

	switch path {
	case "/students":
		return false
	case "/students/register":
		return true
	case "/classes", "/classes/record", "/schedule", "/schedule/record", "/schedule/repeat", "/schedule/series", "/profile", "/logs", "/my-students", "/documents", "/analytics", "/guides", "/learning-materials":
		return true
	default:
		return false
	}
}
