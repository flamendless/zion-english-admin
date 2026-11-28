package constants

import "testing"

func TestValidLeaveRequestType(t *testing.T) {
	if !ValidLeaveRequestType(string(LeaveRequestTypeVacation)) {
		t.Fatal("expected vacation valid")
	}
	if ValidLeaveRequestType("pto") {
		t.Fatal("expected unknown type invalid")
	}
}

func TestValidLeaveRequestStatus(t *testing.T) {
	if !ValidLeaveRequestStatus(string(LeaveRequestStatusSubmitted)) {
		t.Fatal("expected submitted valid")
	}
	if ValidLeaveRequestStatus("pending") {
		t.Fatal("expected pending invalid")
	}
}