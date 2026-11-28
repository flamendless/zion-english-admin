package constants

type LeaveRequestType string

const (
	LeaveRequestTypeSick        LeaveRequestType = "sick"
	LeaveRequestTypeVacation    LeaveRequestType = "vacation"
	LeaveRequestTypeMaternity   LeaveRequestType = "maternity"
	LeaveRequestTypePersonal    LeaveRequestType = "personal"
	LeaveRequestTypeBereavement LeaveRequestType = "bereavement"
	LeaveRequestTypeUnpaid      LeaveRequestType = "unpaid"
)

var LeaveRequestTypes = []LeaveRequestType{
	LeaveRequestTypeSick,
	LeaveRequestTypeVacation,
	LeaveRequestTypeMaternity,
	LeaveRequestTypePersonal,
	LeaveRequestTypeBereavement,
	LeaveRequestTypeUnpaid,
}

func (t LeaveRequestType) Label() string {
	switch t {
	case LeaveRequestTypeSick:
		return "Sick"
	case LeaveRequestTypeVacation:
		return "Vacation"
	case LeaveRequestTypeMaternity:
		return "Maternity"
	case LeaveRequestTypePersonal:
		return "Personal"
	case LeaveRequestTypeBereavement:
		return "Bereavement"
	case LeaveRequestTypeUnpaid:
		return "Unpaid"
	default:
		return string(t)
	}
}

func ValidLeaveRequestType(value string) bool {
	switch LeaveRequestType(value) {
	case LeaveRequestTypeSick,
		LeaveRequestTypeVacation,
		LeaveRequestTypeMaternity,
		LeaveRequestTypePersonal,
		LeaveRequestTypeBereavement,
		LeaveRequestTypeUnpaid:
		return true
	default:
		return false
	}
}

type LeaveRequestStatus string

const (
	LeaveRequestStatusSubmitted LeaveRequestStatus = "submitted"
	LeaveRequestStatusApproved  LeaveRequestStatus = "approved"
	LeaveRequestStatusRejected  LeaveRequestStatus = "rejected"
)

var LeaveRequestStatuses = []LeaveRequestStatus{
	LeaveRequestStatusSubmitted,
	LeaveRequestStatusApproved,
	LeaveRequestStatusRejected,
}

func (s LeaveRequestStatus) Label() string {
	switch s {
	case LeaveRequestStatusSubmitted:
		return "Submitted"
	case LeaveRequestStatusApproved:
		return "Approved"
	case LeaveRequestStatusRejected:
		return "Rejected"
	default:
		return string(s)
	}
}

func ValidLeaveRequestStatus(value string) bool {
	switch LeaveRequestStatus(value) {
	case LeaveRequestStatusSubmitted,
		LeaveRequestStatusApproved,
		LeaveRequestStatusRejected:
		return true
	default:
		return false
	}
}