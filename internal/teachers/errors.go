package teachers

import "errors"

var (
	ErrRolesForbidden      = errors.New("[TEACHERS] you are not allowed to manage roles for this teacher")
	ErrInvalidRole         = errors.New("[TEACHERS] invalid role")
	ErrTeacherRoleRequired = errors.New("[TEACHERS] teacher role is required")
)
