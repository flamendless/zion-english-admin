package notifications

type Category string

const (
	CategorySchedule   Category = "schedule"
	CategoryStudents   Category = "students"
	CategoryDocuments  Category = "documents"
	CategoryPayments   Category = "payments"
	CategoryIntroVideo Category = "intro_video"
	CategoryAccount    Category = "account"
)

type CategoryDefinition struct {
	Category    Category
	Label       string
	Description string
}

func TeacherNotificationCategories() []CategoryDefinition {
	return []CategoryDefinition{
		{
			Category:    CategorySchedule,
			Label:       "Schedule",
			Description: "Changes to your scheduled classes, including edits and cancellations.",
		},
		{
			Category:    CategoryStudents,
			Label:       "Students",
			Description: "Updates when a student is assigned to you or their record changes.",
		},
		{
			Category:    CategoryDocuments,
			Label:       "Documents",
			Description: "Review outcomes for ID documents you submitted.",
		},
		{
			Category:    CategoryPayments,
			Label:       "Payments",
			Description: "Payment sent, receipt confirmed, and received updates.",
		},
		{
			Category:    CategoryIntroVideo,
			Label:       "Intro video",
			Description: "Processing results and review updates for your introduction video.",
		},
		{
			Category:    CategoryAccount,
			Label:       "Account",
			Description: "Approval status and profile updates from administrators.",
		},
	}
}

func CategoryForKind(kind string) (Category, bool) {
	switch kind {
	case KindScheduleChanged:
		return CategorySchedule, true
	case KindStudentUpdated:
		return CategoryStudents, true
	case KindDocumentReviewed:
		return CategoryDocuments, true
	case KindPaymentSent, KindPaymentReceived, KindPaymentConfirmed:
		return CategoryPayments, true
	case KindIntroVideoProcessed, KindIntroVideoFailed, KindIntroVideoReviewed, KindIntroVideoDeleted:
		return CategoryIntroVideo, true
	case KindTeacherApproved, KindProfileUpdated:
		return CategoryAccount, true
	default:
		return "", false
	}
}
