package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestValidateDocumentUploadRejectsEmptyAndOversized(t *testing.T) {
	_, err := validateDocumentUpload(bytes.NewReader(nil), "doc.pdf", 0)
	if err != ErrUploadedFileEmpty {
		t.Fatalf("expected ErrUploadedFileEmpty, got %v", err)
	}

	_, err = validateDocumentUpload(bytes.NewReader(nil), "doc.pdf", maxDocumentBytes+1)
	if err != ErrDocumentFileTooLarge {
		t.Fatalf("expected ErrDocumentFileTooLarge, got %v", err)
	}
}

func TestValidateDocumentUploadAcceptsPDFHeader(t *testing.T) {
	ext, err := validateDocumentUpload(bytes.NewReader([]byte("%PDF-1.4")), "doc.pdf", 8)
	if err != nil || ext != ".pdf" {
		t.Fatalf("expected pdf, got ext=%s err=%v", ext, err)
	}
}

func TestValidateAvatarUploadRejectsUnsupportedFormat(t *testing.T) {
	_, err := validateAvatarUpload(bytes.NewReader([]byte("not-an-image")), 12)
	if err != ErrInvalidAvatarImage {
		t.Fatalf("expected ErrInvalidAvatarImage, got %v", err)
	}
}

func TestValidateIntroVideoUploadRejectsBadExtension(t *testing.T) {
	_, err := validateIntroVideoUpload(bytes.NewReader([]byte("test")), "clip.exe", 4)
	if err != ErrUnsupportedIntroVideoFormat {
		t.Fatalf("expected ErrUnsupportedIntroVideoFormat, got %v", err)
	}
}

func TestValidateIntroVideoUploadAcceptsMP4Header(t *testing.T) {
	header := make([]byte, 12)
	header[4] = 'f'
	header[5] = 't'
	header[6] = 'y'
	header[7] = 'p'
	_, err := validateIntroVideoUpload(bytes.NewReader(header), "clip.mp4", int64(len(header)))
	if err != nil {
		t.Fatalf("expected valid mp4 upload, got %v", err)
	}
}

func TestMaterialLibraryFilters(t *testing.T) {
	rows := []learningMaterialRow{
		{ID: 1, Title: "Grammar", Description: "Basics", Status: "published", Access: "public"},
		{ID: 2, Title: "Vocab", Description: "Words", Status: "draft", Access: "private"},
	}
	tags := map[int64][]int64{
		1: []int64{5},
		2: []int64{7},
	}

	filtered := filterLearningMaterialRows(rows, tags, materialLibraryFilters{
		Query:  "grammar",
		Status: "published",
		TagID:  5,
	})
	if len(filtered) != 1 || filtered[0].ID != 1 {
		t.Fatalf("unexpected filtered rows: %+v", filtered)
	}

	if !matchesMaterialQuery("Hello", "World", "") {
		t.Fatal("empty query should match")
	}
	if matchesMaterialQuery("Hello", "World", "missing") {
		t.Fatal("missing query should not match")
	}
	if !strings.Contains("grammar", "gram") {
		t.Fatal("sanity check")
	}
}
