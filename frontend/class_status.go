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

var StudentFilterStatusOptions = func() []StatusOption {
	opts := make([]StatusOption, 0, len(constants.StudentFilterStatuses))
	for _, status := range constants.StudentFilterStatuses {
		opts = append(opts, StatusOption{
			Value: string(status),
			Label: status.Label(),
		})
	}
	return opts
}()

var TeacherStatusOptions = func() []StatusOption {
	opts := make([]StatusOption, 0, len(constants.TeacherStatuses))
	for _, status := range constants.TeacherStatuses {
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

var TeacherIntroVideoStatusOptions = func() []StatusOption {
	opts := make([]StatusOption, 0, len(constants.TeacherIntroVideoStatuses))
	for _, status := range constants.TeacherIntroVideoStatuses {
		opts = append(opts, StatusOption{
			Value: string(status),
			Label: status.Label(),
		})
	}
	return opts
}()

const TeacherDocsFilterStatusNone = "none"

var TeacherDocsFilterStatusOptions = func() []StatusOption {
	opts := make([]StatusOption, 0, len(TeacherDocumentStatusOptions)+1)
	opts = append(opts, StatusOption{
		Value: TeacherDocsFilterStatusNone,
		Label: "None",
	})
	opts = append(opts, TeacherDocumentStatusOptions...)
	return opts
}()

var TeacherResumeFilterStatusOptions = func() []StatusOption {
	return []StatusOption{
		{
			Value: TeacherDocsFilterStatusNone,
			Label: "None",
		},
		{
			Value: string(constants.TeacherDocumentStatusApproved),
			Label: "Uploaded",
		},
	}
}()

func ResumeStatusDisplayLabel(status constants.TeacherDocumentStatus) string {
	switch status {
	case constants.TeacherDocumentStatusApproved, constants.TeacherDocumentStatusSubmitted:
		return "Uploaded"
	default:
		return ""
	}
}

const (
	TeacherConnectionsFilterZoom           = "zoom"
	TeacherConnectionsFilterGoogleCalendar = "google_calendar"
)

var PaymentMethodOptions = func() []StatusOption {
	opts := make([]StatusOption, 0, len(constants.PaymentMethods))
	for _, method := range constants.PaymentMethods {
		opts = append(opts, StatusOption{
			Value: string(method),
			Label: method.Label(),
		})
	}
	return opts
}()

var PaymentStatusFilterOptions = func() []StatusOption {
	opts := make([]StatusOption, 0, len(constants.PaymentStatuses))
	for _, status := range constants.PaymentStatuses {
		opts = append(opts, StatusOption{
			Value: string(status),
			Label: status.Label(),
		})
	}
	return opts
}()

var CurrencyOptions = func() []StatusOption {
	opts := make([]StatusOption, 0, len(constants.CurrencyCodes))
	for _, code := range constants.CurrencyCodes {
		opts = append(opts, StatusOption{
			Value: code,
			Label: code,
		})
	}
	return opts
}()
