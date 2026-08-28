package frontend

import "zion-english/internal/constants"

type StatusOption struct {
	Value string
	Label string
}

var ClassRecordStatusOptions = func() []StatusOption {
	opts := make([]StatusOption, 0, len(constants.ClassStatuses))
	for _, status := range constants.ClassStatuses {
		opts = append(opts, StatusOption{
			Value: string(status),
			Label: status.Label(),
		})
	}
	return opts
}()

var ClassListFilterOptions = func() []StatusOption {
	opts := make([]StatusOption, 0, len(constants.ClassListFilterStatuses))
	for _, status := range constants.ClassListFilterStatuses {
		opts = append(opts, StatusOption{
			Value: string(status),
			Label: status.Label(),
		})
	}
	return opts
}()

var StudentStatusOptions = func() []StatusOption {
	opts := make([]StatusOption, 0, len(constants.StudentStatuses))
	for _, status := range constants.StudentStatuses {
		opts = append(opts, StatusOption{
			Value: string(status),
			Label: status.Label(),
		})
	}
	return opts
}()

var TeacherFilterStatusOptions = func() []StatusOption {
	opts := make([]StatusOption, 0, len(constants.TeacherFilterStatuses))
	for _, status := range constants.TeacherFilterStatuses {
		opts = append(opts, StatusOption{
			Value: string(status),
			Label: status.Label(),
		})
	}
	return opts
}()

var ClassRecordFilterStatusOptions = func() []StatusOption {
	opts := make([]StatusOption, 0, len(constants.ClassStatuses))
	for _, status := range constants.ClassStatuses {
		opts = append(opts, StatusOption{
			Value: string(status),
			Label: status.Label(),
		})
	}
	return opts
}()

var TeacherDocumentStatusOptions = func() []StatusOption {
	opts := make([]StatusOption, 0, len(constants.TeacherDocumentStatuses))
	for _, status := range constants.TeacherDocumentStatuses {
		opts = append(opts, StatusOption{
			Value: string(status),
			Label: status.Label(),
		})
	}
	return opts
}()
