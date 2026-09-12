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

func (s ClassStatus) Label() string {
	switch s {
	case ClassStatusConducted:
		return "Conducted"
	case ClassStatusCancelled:
		return "Cancelled"
	case ClassStatusRescheduled:
		return "Rescheduled"
	default:
		return string(s)
	}
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

func (s ClassListFilterStatus) Label() string {
	switch s {
	case ClassListFilterScheduled:
		return "Scheduled"
	case ClassListFilterConducted:
		return "Conducted"
	case ClassListFilterCancelled:
		return "Cancelled"
	case ClassListFilterRescheduled:
		return "Rescheduled"
	case ClassListFilterDeleted:
		return "Deleted"
	default:
		return string(s)
	}
}
