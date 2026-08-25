package meetings

import (
	"strings"
	"testing"
)

func TestZoomProviderAuthorizeURL(t *testing.T) {
	p := NewZoomProvider(ZoomConfig{
		ClientID:     "client-id",
		ClientSecret: "secret",
		RedirectURI:  "https://example.com/callback",
		AuthorizeURL: "https://zoom.us/oauth/authorize?response_type=code&client_id=client-id&redirect_uri=https%3A%2F%2Fexample.com%2Fcallback",
	})
	got, err := p.AuthorizeURL("test-state")
	if err != nil {
		t.Fatal(err)
	}
	for _, part := range []string{"state=test-state", "client_id=client-id", "redirect_uri=https%3A%2F%2Fexample.com%2Fcallback"} {
		if !strings.Contains(got, part) {
			t.Fatalf("authorize url missing %q: %s", part, got)
		}
	}
}

func TestZoomProviderAuthorizeURLMissing(t *testing.T) {
	p := NewZoomProvider(ZoomConfig{
		ClientID:     "client-id",
		ClientSecret: "secret",
		RedirectURI:  "https://example.com/callback",
	})
	_, err := p.AuthorizeURL("test-state")
	if err != ErrZoomAuthorizeURLNotSet {
		t.Fatalf("got %v want %v", err, ErrZoomAuthorizeURLNotSet)
	}
}

func TestZoomProviderIsConfiguredRequiresAuthorizeURL(t *testing.T) {
	p := NewZoomProvider(ZoomConfig{
		ClientID:     "client-id",
		ClientSecret: "secret",
		RedirectURI:  "https://example.com/callback",
	})
	if p.IsConfigured() {
		t.Fatal("expected not configured without authorize url")
	}
	p.cfg.AuthorizeURL = "https://zoom.us/oauth/authorize?response_type=code"
	if !p.IsConfigured() {
		t.Fatal("expected configured with authorize url")
	}
}
