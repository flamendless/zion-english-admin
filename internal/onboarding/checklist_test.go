package onboarding

import (
	"testing"

	"zion-english/internal/constants"
)

func TestBuildSummaryComplete(t *testing.T) {
	items := Build(Input{
		TeacherStatus:          constants.TeacherStatusApproved,
		HasProfilePhoto:        true,
		DocsStatus:             string(constants.TeacherDocumentStatusApproved),
		IntroVideoRequired:     true,
		IntroVideoStatus:       string(constants.TeacherIntroVideoStatusApproved),
		TrainingRequiredCompleted: 2,
		TrainingRequiredTotal:   2,
		ZoomShow:               true,
		ZoomConnected:          true,
		GoogleShow:             false,
	})

	completed, total := Summary(items)
	if completed != 6 || total != 6 {
		t.Fatalf("Summary() = %d/%d, want 6/6", completed, total)
	}
	if !IsComplete(items) {
		t.Fatal("expected checklist to be complete")
	}
}

func TestTrainingItemInProgress(t *testing.T) {
	item := trainingItem(1, 3)
	if item.Status != ItemStatusInProgress {
		t.Fatalf("expected in_progress, got %s", item.Status)
	}
	if item.Detail != "Completed 1 of 3" {
		t.Fatalf("unexpected detail: %s", item.Detail)
	}
}

func TestDocumentsRejected(t *testing.T) {
	item := documentsItem(string(constants.TeacherDocumentStatusRejected))
	if item.Status != ItemStatusRejected {
		t.Fatalf("expected rejected, got %s", item.Status)
	}
}

func TestBuildOmitsTrainingWhenNoRequiredPublished(t *testing.T) {
	items := Build(Input{
		TeacherStatus:           constants.TeacherStatusApproved,
		HasProfilePhoto:         true,
		DocsStatus:              string(constants.TeacherDocumentStatusApproved),
		TrainingRequiredTotal:   0,
		TrainingRequiredCompleted: 0,
	})

	for _, item := range items {
		if item.ID == ItemIDTraining {
			t.Fatal("expected training item to be omitted when no required published materials")
		}
	}
}

func TestBuildOmitsIntroVideoWhenNotRequired(t *testing.T) {
	items := Build(Input{
		TeacherStatus:      constants.TeacherStatusApproved,
		HasProfilePhoto:    true,
		DocsStatus:         string(constants.TeacherDocumentStatusApproved),
		IntroVideoRequired: false,
	})

	for _, item := range items {
		if item.ID == ItemIDIntroVideo {
			t.Fatal("expected intro video item to be omitted when not required")
		}
	}
}
