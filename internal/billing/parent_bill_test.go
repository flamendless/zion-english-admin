package billing

import (
	"testing"

	"zion-english/internal/constants"
)

func TestClassRateIsHalfOfTeacherRate(t *testing.T) {
	if !ClassRateIsHalfOfTeacherRate(50, 100) {
		t.Fatal("expected 50 to be half of 100")
	}
	if ClassRateIsHalfOfTeacherRate(0, 100) {
		t.Fatal("zero class rate is not half")
	}
	if ClassRateIsHalfOfTeacherRate(100, 100) {
		t.Fatal("full class rate is not half")
	}
}

func TestEffectiveParentBillAmount(t *testing.T) {
	tests := []struct {
		status      constants.ClassStatus
		parent      float64
		classRate   float64
		teacherRate float64
		want        float64
	}{
		{constants.ClassStatusConducted, 100, 100, 200, 100},
		{constants.ClassStatusCancelled, 100, 50, 100, 50},
		{constants.ClassStatusCancelled, 100, 0, 100, 100},
		{constants.ClassStatusCancelled, 100, 100, 100, 100},
		{constants.ClassStatusRescheduled, 80, 40, 80, 40},
	}
	for _, tt := range tests {
		got := EffectiveParentBillAmount(tt.status, tt.parent, tt.classRate, tt.teacherRate)
		if got != tt.want {
			t.Fatalf("status %s parent %v class %v teacher %v: got %v want %v",
				tt.status, tt.parent, tt.classRate, tt.teacherRate, got, tt.want)
		}
	}
}
