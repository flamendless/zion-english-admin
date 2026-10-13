package frontend

import (
	"testing"

	"zion-english/internal/constants"
)

func TestClassRecordRowData_StatusIconKind(t *testing.T) {
	overdueScheduled := ClassRecordRowData{
		Status:  constants.ClassListFilterScheduled,
		Overdue: true,
	}
	if overdueScheduled.StatusIconKind() != IconKindOverdue {
		t.Fatalf("expected overdue icon for overdue scheduled row")
	}

	if (ClassRecordRowData{Status: constants.ClassListFilterConducted}).StatusIconKind() != IconKindConduct {
		t.Fatal("expected conduct icon for conducted")
	}
	if (ClassRecordRowData{Status: constants.ClassListFilterCancelled}).StatusIconKind() != IconKindCancel {
		t.Fatal("expected cancel icon for cancelled")
	}
}
