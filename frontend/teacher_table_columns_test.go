package frontend

import "testing"

func TestTeacherTableColumnsSummary(t *testing.T) {
	visible := map[TeacherTableColumnID]bool{
		TeacherTableColumnDocsStatus:  true,
		TeacherTableColumnConnections: true,
	}
	if got := teacherTableColumnsSummary(visible); got != "All" {
		t.Fatalf("expected All, got %q", got)
	}

	visible[TeacherTableColumnDocsStatus] = false
	if got := teacherTableColumnsSummary(visible); got != "Connected to" {
		t.Fatalf("expected Connected to, got %q", got)
	}

	visible[TeacherTableColumnConnections] = false
	if got := teacherTableColumnsSummary(visible); got != "None" {
		t.Fatalf("expected None, got %q", got)
	}
}
