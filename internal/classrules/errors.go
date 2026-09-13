package classrules

import "errors"

var (
	ErrInactiveStudent              = errors.New("[CLASS RECORD] cannot record class for an inactive student")
	ErrDuplicateClass               = errors.New("[CLASS RECORD] a class with the same student, teacher, date, and duration already exists")
	ErrStudentNotAssigned           = errors.New("[CLASS RECORD] student is not assigned to this teacher")
	ErrStudentNotFound              = errors.New("[CLASS RECORD] student not found")
	ErrTeacherNotOwner              = errors.New("[CLASS RECORD] you can only edit your own class records")
	ErrAlreadyDeleted               = errors.New("[CLASS RECORD] this class has already been deleted")
	ErrCheckClassDuplicate          = errors.New("[CLASS RECORD] failed to check duplicate class")
	ErrDuplicateScheduled           = errors.New("[SCHEDULE] a scheduled class with the same student, teacher, date, and duration already exists")
	ErrScheduleNotOwner             = errors.New("[SCHEDULE] you can only manage your own scheduled classes")
	ErrTeacherScheduleConflict      = errors.New("[SCHEDULE] teacher already has a class scheduled at this time")
	ErrStudentScheduleConflict      = errors.New("[SCHEDULE] student already has a class scheduled at this time")
	ErrVerifyStudentAssignment      = errors.New("[SCHEDULE] failed to verify student assignment")
	ErrCheckScheduledDuplicate      = errors.New("[SCHEDULE] failed to check duplicate scheduled class")
	ErrCheckClassRecordDuplicate    = errors.New("[SCHEDULE] failed to check duplicate class record")
	ErrCheckTeacherScheduleConflict = errors.New("[SCHEDULE] failed to check teacher schedule conflicts")
	ErrCheckStudentScheduleConflict = errors.New("[SCHEDULE] failed to check student schedule conflicts")
)
