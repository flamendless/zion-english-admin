package constants

type TeacherAdminRoleRequestStatus string

const (
	TeacherAdminRoleRequestStatusPending  TeacherAdminRoleRequestStatus = "pending"
	TeacherAdminRoleRequestStatusApproved TeacherAdminRoleRequestStatus = "approved"
	TeacherAdminRoleRequestStatusRejected TeacherAdminRoleRequestStatus = "rejected"
)

func ValidTeacherAdminRoleRequestStatus(value string) bool {
	switch TeacherAdminRoleRequestStatus(value) {
	case TeacherAdminRoleRequestStatusPending,
		TeacherAdminRoleRequestStatusApproved,
		TeacherAdminRoleRequestStatusRejected:
		return true
	default:
		return false
	}
}
