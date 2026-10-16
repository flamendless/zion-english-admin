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
		status              constants.ClassStatus
		parent              float64
		classRate           float64
		teacherRate         float64
		classDuration       int64
		teacherRateDuration int64
		want                float64
	}{
		{constants.ClassStatusConducted, 100, 100, 200, 25, 25, 100},
		{constants.ClassStatusConducted, 4500, 180, 90, 50, 25, 9000},
		{constants.ClassStatusConducted, 100, 100, 200, 0, 25, 100},
		{constants.ClassStatusCancelled, 100, 50, 100, 25, 25, 50},
		{constants.ClassStatusCancelled, 100, 0, 100, 25, 25, 0},
		{constants.ClassStatusRescheduled, 80, 0, 80, 25, 25, 0},
		{constants.ClassStatusCancelled, 100, 100, 100, 25, 25, 100},
		{constants.ClassStatusRescheduled, 80, 40, 80, 25, 25, 40},
	}
	for _, tt := range tests {
		got := EffectiveParentBillAmount(tt.status, tt.parent, tt.classRate, tt.teacherRate, tt.classDuration, tt.teacherRateDuration)
		if got != tt.want {
			t.Fatalf("status %s parent %v class %v teacher %v dur %v/%v: got %v want %v",
				tt.status, tt.parent, tt.classRate, tt.teacherRate, tt.classDuration, tt.teacherRateDuration, got, tt.want)
		}
	}
}
