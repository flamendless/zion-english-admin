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

func TestAnnouncementsInAdminNavGroup(t *testing.T) {
	group := navGroupForRole(auth.RoleAdmin, "admin")
	if group == nil {
		t.Fatal("admin nav group expected")
	}
	if len(group.Items) == 0 {
		t.Fatal("admin nav group should have items")
	}
	last := group.Items[len(group.Items)-1]
	if last.Path != "/announcements" {
		t.Fatalf("expected announcements last in admin group, got %s", last.Path)
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
