package ads

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"zion-english/internal/constants"
)

func TestClearLegacyRotationCookiesClearsAll(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/", nil)
	const n = 50
	for i := 0; i < n; i++ {
		name := cookieNameForAdZone(int64(i+1), constants.AdZoneTop)
		req.AddCookie(&http.Cookie{Name: name, Value: "1"})
	}
	rec := httptest.NewRecorder()
	clearLegacyRotationCookies(rec, req, "/", 0)
	cleared := 0
	for _, c := range rec.Result().Cookies() {
		if c.MaxAge == -1 && strings.HasPrefix(c.Name, cookiePrefix) {
			cleared++
		}
	}
	if cleared != n {
		t.Fatalf("cleared %d legacy cookies, want %d", cleared, n)
	}
}

func TestClearLegacyRotationCookiesRespectsPartialLimit(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/zion-english-admin/classes/partials/rows", nil)
	const n = 40
	for i := 0; i < n; i++ {
		name := cookieNameForAdZone(int64(i+1), constants.AdZoneTop)
		req.AddCookie(&http.Cookie{Name: name, Value: "1"})
	}
	rec := httptest.NewRecorder()
	clearLegacyRotationCookies(rec, req, "/zion-english-admin", maxLegacyRotationClearsPerPartialResponse)
	cleared := 0
	for _, c := range rec.Result().Cookies() {
		if c.MaxAge == -1 && strings.HasPrefix(c.Name, cookiePrefix) {
			cleared++
		}
	}
	if cleared != maxLegacyRotationClearsPerPartialResponse {
		t.Fatalf("cleared %d legacy cookies, want %d", cleared, maxLegacyRotationClearsPerPartialResponse)
	}
}

func TestPurgeLegacyUsesClearSiteDataOnLoginWithManyLegacyCookies(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/zion-english-admin/auth/login", nil)
	const n = maxLegacyRotationClearsPerPartialResponse + 5
	for i := 0; i < n; i++ {
		name := cookieNameForAdZone(int64(i+1), constants.AdZoneTop)
		req.AddCookie(&http.Cookie{Name: name, Value: "1"})
	}
	rec := httptest.NewRecorder()
	PurgeLegacyRotationCookies(rec, req)
	if got := rec.Header().Get("Clear-Site-Data"); got != `"cookies"` {
		t.Fatalf("Clear-Site-Data = %q, want %q", got, `"cookies"`)
	}
	if len(rec.Result().Cookies()) != 0 {
		t.Fatalf("expected no Set-Cookie clears, got %d", len(rec.Result().Cookies()))
	}
}

func TestNeedsRotationCookiePurgeDetectsLegacyPicks(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: rotationPicksCookieName, Value: "bloated"})
	if !needsRotationCookiePurge(req) {
		t.Fatal("expected purge when zion_ad_picks present")
	}
}
