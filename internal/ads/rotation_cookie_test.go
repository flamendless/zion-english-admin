package ads

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
