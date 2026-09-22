package cmd

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIsTrainingMaterialWatchRequest(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{path: "/zion-english-admin/training-materials/12/watch", want: true},
		{path: "/zion-english-admin/training-materials/12/watch/", want: true},
		{path: "/training-materials/12/watch", want: true},
		{path: "/zion-english-admin/training-materials/12/progress", want: false},
		{path: "/zion-english-admin/training-materials", want: false},
		{path: "/zion-english-admin/learning-materials/12/watch", want: false},
	}

	for _, tc := range tests {
		req := httptest.NewRequest("GET", tc.path, nil)
		if got := isTrainingMaterialWatchRequest(req); got != tc.want {
			t.Fatalf("isTrainingMaterialWatchRequest(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestContentSecurityPolicyTrainingWatchAllowsYouTube(t *testing.T) {
	req := httptest.NewRequest("GET", "/zion-english-admin/training-materials/1/watch", nil)
	csp := contentSecurityPolicy(req)
	if !strings.Contains(csp, "script-src 'self' 'unsafe-inline' https://www.youtube.com") {
		t.Fatalf("missing youtube script-src in CSP: %q", csp)
	}
	if !strings.Contains(csp, "frame-src 'self' https://www.youtube.com https://www.youtube-nocookie.com") {
		t.Fatalf("missing youtube frame-src in CSP: %q", csp)
	}

	defaultReq := httptest.NewRequest("GET", "/zion-english-admin/dashboard", nil)
	defaultCSP := contentSecurityPolicy(defaultReq)
	if defaultCSP == csp {
		t.Fatalf("expected default CSP to differ from training watch CSP")
	}
}
