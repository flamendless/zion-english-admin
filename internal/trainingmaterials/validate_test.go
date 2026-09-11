package trainingmaterials

import "testing"

func TestValidateRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     Request
		wantErr error
	}{
		{
			name: "valid published",
			req: Request{
				Title:       "Onboarding video",
				Description: "Intro for new teachers",
				URL:         "https://www.youtube.com/watch?v=abc123",
				Status:      string(StatusPublished),
			},
		},
		{
			name: "missing title",
			req: Request{
				URL:    "https://drive.google.com/file/d/abc/view",
				Status: string(StatusDraft),
			},
			wantErr: ErrTitleRequired,
		},
		{
			name: "invalid url host",
			req: Request{
				Title:  "Slides",
				URL:    "https://example.com/file",
				Status: string(StatusDraft),
			},
			wantErr: ErrInvalidURL,
		},
		{
			name: "invalid status",
			req: Request{
				Title:  "Slides",
				URL:    "https://docs.google.com/presentation/d/abc/edit",
				Status: string(StatusDeleted),
			},
			wantErr: ErrInvalidStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequest(tt.req)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("ValidateRequest() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ValidateRequest() unexpected error: %v", err)
			}
		})
	}
}

func TestIsAllowedURL(t *testing.T) {
	allowed := []string{
		"https://www.youtube.com/watch?v=abc",
		"https://youtu.be/abc",
		"https://m.youtube.com/watch?v=abc",
		"https://drive.google.com/file/d/abc/view",
		"https://docs.google.com/presentation/d/abc/edit",
	}
	for _, raw := range allowed {
		if !IsAllowedURL(raw) {
			t.Fatalf("expected allowed url: %s", raw)
		}
	}

	disallowed := []string{
		"https://example.com/video",
		"ftp://drive.google.com/file",
		"not-a-url",
	}
	for _, raw := range disallowed {
		if IsAllowedURL(raw) {
			t.Fatalf("expected disallowed url: %s", raw)
		}
	}
}
