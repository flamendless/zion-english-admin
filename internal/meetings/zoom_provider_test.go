package meetings

import (
	"strings"
	"testing"
)

func TestZoomProviderAuthorizeURLDefault(t *testing.T) {
	p := NewZoomProvider(ZoomConfig{
		ClientID:     "client-id",
		ClientSecret: "secret",
		RedirectURI:  "https://example.com/callback",
	})
	got, err := p.AuthorizeURL("test-state")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://zoom.us/oauth/authorize?client_id=client-id&redirect_uri=https%3A%2F%2Fexample.com%2Fcallback&response_type=code&state=test-state"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestZoomProviderAuthorizeURLBetaOverride(t *testing.T) {
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
