package ads

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"zion-english/internal/constants"
)

func TestPersistRotationPicksConsolidatesLegacyCookies(t *testing.T) {
	t.Parallel()
	legacyName := cookieNameForAdZone(1, constants.AdZoneTop)
	req := httptest.NewRequest("GET", "/schedule", nil)
	req.AddCookie(&http.Cookie{Name: legacyName, Value: "42"})
	rec := httptest.NewRecorder()

	persistRotationPicks(rec, req, "/", map[string]int64{legacyName: 42})

	res := rec.Result()
	var consolidated bool
	for _, c := range res.Cookies() {
		if c.Name == rotationPicksCookieName {
			consolidated = true
			picks := make(map[string]int64)
			decodeRotationPicksCookie(c.Value, picks)
			if picks[legacyName] != 42 {
				t.Fatalf("consolidated cookie missing legacy key: %v", picks)
			}
		}
		if c.Name == legacyName && c.MaxAge != -1 {
			t.Fatalf("expected legacy cookie cleared, got MaxAge=%d", c.MaxAge)
		}
	}
	if !consolidated {
		t.Fatal("expected consolidated rotation cookie")
	}
}

func TestPersistRotationPicksDoesNotImportLegacyCookieNames(t *testing.T) {
	t.Parallel()
	legacyName := cookieNameForAdZone(3, constants.AdZoneTop)
	req := httptest.NewRequest("GET", "/schedule", nil)
	req.AddCookie(&http.Cookie{Name: legacyName, Value: "42"})
	rec := httptest.NewRecorder()

	persistRotationPicks(rec, req, "/", nil)

	for _, c := range rec.Result().Cookies() {
		if c.Name != rotationPicksCookieName {
			continue
		}
		picks := make(map[string]int64)
		decodeRotationPicksCookie(c.Value, picks)
		if len(picks) != 0 {
			t.Fatalf("expected empty consolidated picks after legacy-only migrate, got %v", picks)
		}
	}
}

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

func TestPurgeLegacyUsesClearSiteDataOnLandingWithManyLegacyCookies(t *testing.T) {
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

func TestRotationPicksFromRequestPrefersConsolidatedCookie(t *testing.T) {
	t.Parallel()
	legacyName := cookieNameForAdZone(2, constants.AdZoneLeft)
	req := httptest.NewRequest("GET", "/", nil)
	raw, _ := json.Marshal(map[string]int64{legacyName: 99})
	encoded := base64.RawURLEncoding.EncodeToString(raw)
	req.AddCookie(&http.Cookie{Name: rotationPicksCookieName, Value: encoded})
	picks := rotationPicksFromRequest(req)
	if picks[legacyName] != 99 {
		t.Fatalf("pick = %d, want 99", picks[legacyName])
	}
}
