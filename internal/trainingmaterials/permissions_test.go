package trainingmaterials

import (
	"testing"

	"zion-english/internal/auth"
)

func TestTrainingMaterialPermissions(t *testing.T) {
	admin := auth.RoleAdmin
	teacher := auth.RoleTeacher
	tester := auth.RoleTester

	if !CanCreate(admin) || CanCreate(teacher) {
		t.Fatal("only admins should create training materials")
	}
	if !CanView(admin, StatusDraft) || CanView(teacher, StatusDraft) {
		t.Fatal("only admins should view draft training materials")
	}
	if !CanView(teacher, StatusPublished) {
		t.Fatal("teacher should view published training materials")
	}
	if CanWatch(tester, StatusPublished) {
		t.Fatal("tester should not watch training materials")
	}
	if !CanWatch(teacher, StatusPublished) {
		t.Fatal("teacher should watch published training materials")
	}
}
