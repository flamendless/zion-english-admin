package constants

type ClassStatus string

const (
	ClassStatusConducted   ClassStatus = "conducted"
	ClassStatusCancelled   ClassStatus = "cancelled"
	ClassStatusRescheduled ClassStatus = "rescheduled"
)

var ClassStatuses = []ClassStatus{
	ClassStatusConducted,
	ClassStatusCancelled,
	ClassStatusRescheduled,
}

type ClassListFilterStatus string

const (
	ClassListFilterScheduled   ClassListFilterStatus = "scheduled"
	ClassListFilterConducted   ClassListFilterStatus = "conducted"
	ClassListFilterCancelled   ClassListFilterStatus = "cancelled"
	ClassListFilterRescheduled ClassListFilterStatus = "rescheduled"
	ClassListFilterDeleted     ClassListFilterStatus = "deleted"
)

var ClassListFilterStatuses = []ClassListFilterStatus{
	ClassListFilterScheduled,
	ClassListFilterConducted,
	ClassListFilterCancelled,
	ClassListFilterRescheduled,
	ClassListFilterDeleted,
}
