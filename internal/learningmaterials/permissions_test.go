package learningmaterials

import (
	"testing"

	"zion-english/internal/auth"
)

func TestCanView(t *testing.T) {
	owner := auth.User{ID: 10, Role: auth.RoleTeacher}
	other := auth.User{ID: 20, Role: auth.RoleTeacher}
	superuser := auth.User{ID: 0, Role: auth.RoleSuperuser}
	admin := auth.User{ID: 1, Role: auth.RoleAdmin}

	if !CanView(superuser, 10, StatusDraft, AccessPrivate) {
		t.Fatal("superuser should view any material")
	}
	if CanView(admin, 10, StatusDraft, AccessPrivate) {
		t.Fatal("admin should follow the same visibility rules as teachers")
	}
	if !CanView(owner, 10, StatusDraft, AccessPrivate) {
		t.Fatal("owner should view own material")
	}
	if CanView(other, 10, StatusDraft, AccessPrivate) {
		t.Fatal("other teacher should not view private draft")
	}
	if !CanView(other, 10, StatusPublished, AccessPublic) {
		t.Fatal("other teacher should view published public material")
	}
	if CanView(other, 10, StatusDeleted, AccessPublic) {
		t.Fatal("deleted material should not be visible")
	}
}

func TestCanEditAndDelete(t *testing.T) {
	owner := auth.User{ID: 10, Role: auth.RoleTeacher}
	other := auth.User{ID: 20, Role: auth.RoleTeacher}
	superuser := auth.User{ID: 0, Role: auth.RoleSuperuser}

	if !CanEdit(owner, 10, StatusDraft) {
		t.Fatal("owner should edit own material")
	}
	if CanEdit(other, 10, StatusDraft) {
		t.Fatal("other teacher should not edit")
	}
	if !CanDelete(superuser) {
		t.Fatal("superuser should delete")
	}
	if CanDelete(owner) {
		t.Fatal("teacher should not delete")
	}
}
