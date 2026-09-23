package onboarding

import (
	"fmt"

	"zion-english/internal/constants"
)

type ItemStatus string

const (
	ItemStatusComplete    ItemStatus = "complete"
	ItemStatusInProgress  ItemStatus = "in_progress"
	ItemStatusNotStarted  ItemStatus = "not_started"
	ItemStatusRejected    ItemStatus = "rejected"
	ItemStatusNotRequired ItemStatus = "not_required"
)

type ItemID string

const (
	ItemIDAccount         ItemID = "account"
	ItemIDProfilePhoto    ItemID = "profile_photo"
	ItemIDDocuments       ItemID = "documents"
	ItemIDResume          ItemID = "resume"
	ItemIDIntroVideo      ItemID = "intro_video"
	ItemIDTraining        ItemID = "training"
	ItemIDZoom            ItemID = "zoom"
	ItemIDGoogleCalendar  ItemID = "google_calendar"
)

type Item struct {
	ID         ItemID
	Label      string
	Status     ItemStatus
	Detail     string
	ActionPath string
}

type Input struct {
	TeacherStatus          constants.TeacherStatus
	HasProfilePhoto        bool
	DocsStatus             string
	ResumeStatus           string
	IntroVideoStatus       string
	IntroVideoRequired     bool
	TrainingRequiredCompleted int64
	TrainingRequiredTotal   int64
	ZoomShow               bool
	ZoomConnected          bool
	GoogleShow             bool
	GoogleConnected        bool
}

func Build(input Input) []Item {
	items := []Item{
		accountItem(input.TeacherStatus),
		profilePhotoItem(input.HasProfilePhoto),
		documentsItem(input.DocsStatus),
		resumeItem(input.ResumeStatus),
	}

	if input.IntroVideoRequired {
		items = append(items, introVideoItem(input.IntroVideoStatus))
	}

	if input.TrainingRequiredTotal > 0 {
		items = append(items, trainingItem(input.TrainingRequiredCompleted, input.TrainingRequiredTotal))
	}

	if input.ZoomShow {
		items = append(items, zoomItem(input.ZoomConnected))
	}
	if input.GoogleShow {
		items = append(items, googleCalendarItem(input.GoogleConnected))
	}

	return items
}

func Summary(items []Item) (completed int, total int) {
	for _, item := range items {
		if item.Status == ItemStatusNotRequired {
			continue
		}
		total++
		if item.Status == ItemStatusComplete {
			completed++
		}
	}
	return completed, total
}

func IsComplete(items []Item) bool {
	completed, total := Summary(items)
	return total > 0 && completed == total
}

func accountItem(status constants.TeacherStatus) Item {
	item := Item{
		ID:         ItemIDAccount,
		Label:      "Account approved",
		ActionPath: "/profile",
	}
	switch status {
	case constants.TeacherStatusApproved:
		item.Status = ItemStatusComplete
		item.Detail = "Approved"
	case constants.TeacherStatusPending:
		item.Status = ItemStatusInProgress
		item.Detail = "Pending approval"
	case constants.TeacherStatusResigned:
		item.Status = ItemStatusNotStarted
		item.Detail = "Resigned"
	default:
		item.Status = ItemStatusNotStarted
		item.Detail = "Not approved"
	}
	return item
}

func profilePhotoItem(hasPhoto bool) Item {
	item := Item{
		ID:         ItemIDProfilePhoto,
		Label:      "Profile photo",
		ActionPath: "/profile",
	}
	if hasPhoto {
		item.Status = ItemStatusComplete
		item.Detail = "Uploaded"
	} else {
		item.Status = ItemStatusNotStarted
		item.Detail = "Upload a profile photo"
	}
	return item
}

func documentsItem(status string) Item {
	item := Item{
		ID:         ItemIDDocuments,
		Label:      "ID documents",
		ActionPath: "/documents",
	}
	switch constants.TeacherDocumentStatus(status) {
	case constants.TeacherDocumentStatusApproved:
		item.Status = ItemStatusComplete
		item.Detail = "Approved"
	case constants.TeacherDocumentStatusSubmitted:
		item.Status = ItemStatusInProgress
		item.Detail = "Submitted, awaiting review"
	case constants.TeacherDocumentStatusRejected:
		item.Status = ItemStatusRejected
		item.Detail = "Rejected, re-upload required"
	default:
		item.Status = ItemStatusNotStarted
		item.Detail = "Upload your ID documents"
	}
	return item
}

func resumeItem(status string) Item {
	item := Item{
		ID:         ItemIDResume,
		Label:      "Resume/CV",
		ActionPath: "/profile",
	}
	switch constants.TeacherDocumentStatus(status) {
	case constants.TeacherDocumentStatusApproved, constants.TeacherDocumentStatusSubmitted:
		item.Status = ItemStatusComplete
		item.Detail = "Uploaded"
	default:
		item.Status = ItemStatusNotStarted
		item.Detail = "Upload your resume/CV"
	}
	return item
}

func introVideoItem(status string) Item {
	item := Item{
		ID:         ItemIDIntroVideo,
		Label:      "Intro video",
		ActionPath: "/intro-videos",
	}
	switch constants.TeacherIntroVideoStatus(status) {
	case constants.TeacherIntroVideoStatusApproved:
		item.Status = ItemStatusComplete
		item.Detail = "Approved"
	case constants.TeacherIntroVideoStatusSubmitted:
		item.Status = ItemStatusInProgress
		item.Detail = "Submitted, awaiting review"
	case constants.TeacherIntroVideoStatusRejected:
		item.Status = ItemStatusRejected
		item.Detail = "Rejected, re-upload required"
	default:
		item.Status = ItemStatusNotStarted
		item.Detail = "Upload your introduction video"
	}
	return item
}

func trainingItem(completed, total int64) Item {
	item := Item{
		ID:         ItemIDTraining,
		Label:      "Training materials",
		ActionPath: "/training-materials",
	}
	if total <= 0 {
		item.Status = ItemStatusNotRequired
		item.Detail = "No required training yet"
		return item
	}
	if completed >= total {
		item.Status = ItemStatusComplete
		item.Detail = "All required training completed"
		return item
	}
	if completed > 0 {
		item.Status = ItemStatusInProgress
		item.Detail = "Completed " + formatCount(completed) + " of " + formatCount(total)
		return item
	}
	item.Status = ItemStatusNotStarted
	item.Detail = "Complete " + formatCount(total) + " required training videos"
	return item
}

func zoomItem(connected bool) Item {
	item := Item{
		ID:         ItemIDZoom,
		Label:      "Zoom connected",
		ActionPath: "/profile",
	}
	if connected {
		item.Status = ItemStatusComplete
		item.Detail = "Connected"
	} else {
		item.Status = ItemStatusNotStarted
		item.Detail = "Connect Zoom on your profile"
	}
	return item
}

func googleCalendarItem(connected bool) Item {
	item := Item{
		ID:         ItemIDGoogleCalendar,
		Label:      "Google Calendar connected",
		ActionPath: "/profile",
	}
	if connected {
		item.Status = ItemStatusComplete
		item.Detail = "Connected"
	} else {
		item.Status = ItemStatusNotStarted
		item.Detail = "Connect Google Calendar on your profile"
	}
	return item
}

func formatCount(n int64) string {
	return fmt.Sprintf("%d", n)
}
