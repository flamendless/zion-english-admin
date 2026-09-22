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

func TestIsNavAccessibleTeacherResources(t *testing.T) {
	for _, path := range []string{"/learning-materials", "/training-materials", "/documents", "/intro-videos"} {
		if !IsNavAccessible(auth.RoleTeacher, path) {
			t.Fatalf("teacher should access %s", path)
		}
	}
}
