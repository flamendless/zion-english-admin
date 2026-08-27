package frontend

import "zion-english/internal/constants"

type StatusOption struct {
	Value string
	Label string
}

func classStatusLabel(status constants.ClassStatus) string {
	switch status {
	case constants.ClassStatusConducted:
		return "Conducted"
	case constants.ClassStatusCancelled:
		return "Cancelled"
	case constants.ClassStatusRescheduled:
		return "Rescheduled"
	default:
		return string(status)
	}
}

func classListFilterStatusLabel(status constants.ClassListFilterStatus) string {
	switch status {
	case constants.ClassListFilterScheduled:
		return "Scheduled"
	case constants.ClassListFilterConducted:
		return "Conducted"
	case constants.ClassListFilterCancelled:
		return "Cancelled"
	case constants.ClassListFilterRescheduled:
		return "Rescheduled"
	case constants.ClassListFilterDeleted:
		return "Deleted"
	default:
		return string(status)
	}
}

var ClassRecordStatusOptions = func() []StatusOption {
	opts := make([]StatusOption, 0, len(constants.ClassStatuses))
	for _, status := range constants.ClassStatuses {
		opts = append(opts, StatusOption{
			Value: string(status),
			Label: classStatusLabel(status),
		})
	}
	return opts
}()

var ClassListFilterOptions = func() []StatusOption {
	opts := make([]StatusOption, 0, len(constants.ClassListFilterStatuses))
	for _, status := range constants.ClassListFilterStatuses {
		opts = append(opts, StatusOption{
			Value: string(status),
			Label: classListFilterStatusLabel(status),
		})
	}
	return opts
}()
