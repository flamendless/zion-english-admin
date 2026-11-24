package constants

import "testing"

func TestPayrollTeacherStatusSortOrder(t *testing.T) {
	if PayrollTeacherStatusAccomplishDailyLogs.SortOrder() >= PayrollTeacherStatusDoneDailyLogs.SortOrder() {
		t.Fatal("accomplish_daily_logs should sort before done_daily_logs")
	}
	if PayrollTeacherStatusSkip.SortOrder() <= PayrollTeacherStatusSalaryReceived.SortOrder() {
		t.Fatal("skip should sort after workflow statuses")
	}
}

func TestPayrollTeacherStatusCountsTowardSessionComplete(t *testing.T) {
	if !PayrollTeacherStatusSalaryReceived.CountsTowardSessionComplete() {
		t.Fatal("salary_received should count toward session complete")
	}
	if !PayrollTeacherStatusSkip.CountsTowardSessionComplete() {
		t.Fatal("skip should count toward session complete")
	}
	if PayrollTeacherStatusAudited.CountsTowardSessionComplete() {
		t.Fatal("audited should not count toward session complete")
	}
}

func TestValidPayrollTeacherStatusSkip(t *testing.T) {
	if !ValidPayrollTeacherStatus("skip") {
		t.Fatal("ValidPayrollTeacherStatus(skip) = false, want true")
	}
}
