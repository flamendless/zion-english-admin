package frontend

import (
	"testing"

	"zion-english/internal/auth"
)

func TestIsNavAccessibleTesterSandbox(t *testing.T) {
	allowed := []string{
		"/profile",
		"/classes",
		"/schedule",
		"/schedule/series",
	}
	for _, path := range allowed {
		if !IsNavAccessible(auth.RoleTester, path) {
			t.Fatalf("tester should access %s", path)
		}
	}

	blocked := []string{
		"/learning-materials",
		"/training-materials",
		"/documents",
		"/guides",
	}
	for _, path := range blocked {
		if IsNavAccessible(auth.RoleTester, path) {
			t.Fatalf("tester should not access %s via nav", path)
		}
	}
}

func TestIsNavAccessibleSuperuserProfile(t *testing.T) {
	if IsNavAccessible(auth.RoleSuperuser, "/profile") {
		t.Fatal("superuser should not access /profile via nav")
	}
	if !IsNavAccessible(auth.RoleAdmin, "/profile") {
		t.Fatal("admin should access /profile")
	}
	if !IsNavAccessible(auth.RoleTeacher, "/profile") {
		t.Fatal("teacher should access /profile")
	}
}

func TestIsNavAccessibleTeacherResources(t *testing.T) {
	for _, path := range []string{"/learning-materials", "/training-materials", "/documents", "/intro-videos"} {
		if !IsNavAccessible(auth.RoleTeacher, path) {
			t.Fatalf("teacher should access %s", path)
		}
	}
}

func TestTeacherDashboardQuickActionsMyHubBeforeProfile(t *testing.T) {
	entries := DashboardQuickActionEntries(auth.RoleTeacher)
	var sawHub bool
	for _, entry := range entries {
		if entry.Kind != NavBarEntryLink {
			continue
		}
		if entry.Item.Path == "/hub" {
			sawHub = true
			break
		}
		if entry.Item.Path == "/profile" {
			t.Fatal("My Hub should appear before My Profile in dashboard quick actions")
		}
	}
	if !sawHub {
		t.Fatal("teacher dashboard quick actions should include My Hub")
	}
}

func TestTeacherDashboardQuickActionsTimeOffRequests(t *testing.T) {
	entries := DashboardQuickActionEntries(auth.RoleTeacher)
	var sawClasses bool
	for _, entry := range entries {
		if entry.Kind == NavBarEntryGroup && entry.Group.ID == "classes" {
			sawClasses = true
			continue
		}
		if entry.Kind != NavBarEntryLink {
			continue
		}
		if entry.Item.Path == "/leave-requests" {
			if !sawClasses {
				t.Fatal("Time-off Requests should appear after the Classes quick action card")
			}
			if entry.Item.Title != "Time-off Requests" {
				t.Fatalf("time-off quick action title = %q, want Time-off Requests", entry.Item.Title)
			}
			if !entry.Item.FeatureCard {
				t.Fatal("time-off quick action should be a feature card")
			}
			return
		}
	}
	t.Fatal("teacher dashboard quick actions should include Time-off Requests")
}

func TestIsNavAccessibleHub(t *testing.T) {
	if !IsNavAccessible(auth.RoleTeacher, "/hub") {
		t.Fatal("teacher should access /hub")
	}
	if IsNavAccessible(auth.RoleAdmin, "/hub") {
		t.Fatal("admin should not see /hub in nav")
	}
	if IsNavAccessible(auth.RoleSuperuser, "/hub") {
		t.Fatal("superuser should not see /hub in nav")
	}
}

func TestAnnouncementsInAdminNavGroup(t *testing.T) {
	group := navGroupForRole(auth.RoleAdmin, "admin")
	if group == nil {
		t.Fatal("admin nav group expected")
	}
	if len(group.Items) == 0 {
		t.Fatal("admin nav group should have items")
	}
	last := group.Items[len(group.Items)-1]
	if last.Path != "/leave-requests" {
		t.Fatalf("expected time-offs last in admin group, got %s", last.Path)
	}
	if last.Title != "Time-offs" {
		t.Fatalf("admin group time-offs title = %q, want Time-offs", last.Title)
	}
	for _, item := range group.Items {
		if item.Path == "/process" {
			t.Fatal("process should not be in admin nav group")
		}
		if item.Path == "/feature-flags" || item.Path == "/logs" {
			t.Fatalf("unexpected %s in admin group", item.Path)
		}
	}
	for _, item := range DashboardExtraItems(auth.RoleAdmin) {
		if item.Path == "/announcements" {
			t.Fatal("announcements should not be a standalone dashboard card")
		}
	}
	if !IsNavAccessible(auth.RoleAdmin, "/announcements") {
		t.Fatal("admin should access /announcements")
	}
}

func TestIsNavAccessibleAdminPayroll(t *testing.T) {
	if !IsNavAccessible(auth.RoleAdmin, "/payroll") {
		t.Fatal("admin should access /payroll")
	}
	if IsNavAccessible(auth.RoleTeacher, "/payroll") {
		t.Fatal("teacher should not access /payroll via nav")
	}
}

func TestIsNavAccessibleParentPaymentOCR(t *testing.T) {
	if !IsNavAccessible(auth.RoleAdmin, "/parent-payment-ocr") {
		t.Fatal("admin should access /parent-payment-ocr")
	}
	if !IsNavAccessible(auth.RoleSuperuser, "/parent-payment-ocr") {
		t.Fatal("superuser should access /parent-payment-ocr")
	}
	if IsNavAccessible(auth.RoleTeacher, "/parent-payment-ocr") {
		t.Fatal("teacher should not access /parent-payment-ocr via nav")
	}
}

func TestLogsNavGroup(t *testing.T) {
	group := navGroupForRole(auth.RoleAdmin, "logs")
	if group == nil {
		t.Fatal("logs nav group expected for admin")
	}
	if group.Label != "Logs" {
		t.Fatalf("logs group label = %q, want Logs", group.Label)
	}
	var sawAudit, sawUpload bool
	for _, item := range group.Items {
		switch item.Path {
		case "/logs":
			sawAudit = true
			if item.Title != "Audit Logs" {
				t.Fatalf("logs item title = %q, want Audit Logs", item.Title)
			}
		case "/upload-logs":
			sawUpload = true
		}
	}
	if !sawAudit || !sawUpload {
		t.Fatal("logs group should include audit logs and upload logs")
	}
}

func TestSuperuserNavGroup(t *testing.T) {
	group := navGroupForRole(auth.RoleSuperuser, "superuser")
	if group == nil {
		t.Fatal("superuser nav group expected")
	}
	if !group.SuperuserOnlyCard {
		t.Fatal("superuser group should use SuperuserOnlyCard")
	}
	var sawFeatureFlags bool
	for _, item := range group.Items {
		if item.Path == "/feature-flags" {
			sawFeatureFlags = true
		}
	}
	if !sawFeatureFlags {
		t.Fatal("superuser group should include feature flags")
	}
	var sawSuperuserCard bool
	for _, entry := range DashboardQuickActionEntries(auth.RoleAdmin) {
		if entry.Kind == NavBarEntryGroup && entry.Group.ID == "superuser" {
			sawSuperuserCard = true
			break
		}
	}
	if sawSuperuserCard {
		t.Fatal("admin dashboard should not show superuser quick action card")
	}
}

func TestProcessHiddenFromNavBarItems(t *testing.T) {
	for _, item := range NavBarItems(auth.RoleAdmin) {
		if item.Path == "/process" {
			t.Fatal("process should be hidden from navbar items")
		}
	}
	if !IsNavAccessible(auth.RoleAdmin, "/process") {
		t.Fatal("admin should still access /process by URL")
	}
}

func TestIsNavAccessibleSuperuserAccesses(t *testing.T) {
	if !IsNavAccessible(auth.RoleSuperuser, "/accesses") {
		t.Fatal("superuser should access /accesses via nav")
	}
	if IsNavAccessible(auth.RoleAdmin, "/accesses") {
		t.Fatal("admin should not access /accesses via nav")
	}
	if IsNavAccessible(auth.RoleTeacher, "/accesses") {
		t.Fatal("teacher should not access /accesses via nav")
	}
}

func TestNavAccessMatrixRows_MatchesIsNavAccessible(t *testing.T) {
	for _, row := range NavAccessMatrixRows() {
		if !navAccessMatrixRowMatchesIsNavAccessible(row) {
			t.Fatalf("matrix row %s out of sync with IsNavAccessible", row.Path)
		}
	}
	spotCheck := map[string]struct {
		teacher bool
		admin   bool
	}{
		"/hub":      {teacher: true, admin: false},
		"/settings": {teacher: false, admin: false},
		"/payroll":  {teacher: false, admin: true},
	}
	for _, row := range NavAccessMatrixRows() {
		want, ok := spotCheck[row.Path]
		if !ok {
			continue
		}
		if row.Teacher != want.teacher || row.Admin != want.admin {
			t.Fatalf("matrix spot check failed for %s: teacher=%v admin=%v", row.Path, row.Teacher, row.Admin)
		}
	}
}
