package frontend

import "testing"

func TestTeacherTableColumnsSummary(t *testing.T) {
	if got := teacherTableColumnsDefaultSummary(); got != TeacherTableColumnsSummaryNone {
		t.Fatalf("expected default None, got %q", got)
	}

	visible := map[TeacherTableColumnID]bool{
		TeacherTableColumnDocsStatus:    true,
		TeacherTableColumnResumeStatus:  true,
		TeacherTableColumnConnections:   true,
	}
	if got := teacherTableColumnsSummary(visible); got != TeacherTableColumnsSummaryAll {
		t.Fatalf("expected All, got %q", got)
	}

	visible[TeacherTableColumnDocsStatus] = false
	visible[TeacherTableColumnResumeStatus] = false
	if got := teacherTableColumnsSummary(visible); got != "Connected to" {
		t.Fatalf("expected Connected to, got %q", got)
	}

	visible[TeacherTableColumnConnections] = false
	if got := teacherTableColumnsSummary(visible); got != TeacherTableColumnsSummaryNone {
		t.Fatalf("expected None, got %q", got)
	}
}
