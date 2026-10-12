package ads

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"zion-english/internal/conf"
)

const rotationPicksCookieName = "zion_ad_picks"

// maxRotationCookieBytes keeps the consolidated cookie under typical 4KB browser limits
// with room for other session cookies in the same header.
const maxRotationCookieBytes = 3800

// maxLegacyRotationClearsPerResponse avoids oversized response Set-Cookie bursts on first migrate.
const maxLegacyRotationClearsPerResponse = 40

type requestRotationPicksKeyType struct{}

var requestRotationPicksKey = requestRotationPicksKeyType{}

func withRequestRotationPicks(ctx context.Context, r *http.Request) context.Context {
	picks := rotationPicksFromRequest(r)
	return context.WithValue(ctx, requestRotationPicksKey, picks)
}

func rotationPicksFromRequest(r *http.Request) map[string]int64 {
	out := make(map[string]int64)
	if r == nil {
		return out
	}
	if c, err := r.Cookie(rotationPicksCookieName); err == nil && c.Value != "" {
		decodeRotationPicksCookie(c.Value, out)
	}
	for _, c := range r.Cookies() {
		if c.Name == rotationPicksCookieName {
			continue
		}
		if !strings.HasPrefix(c.Name, cookiePrefix) {
			continue
		}
		if id, err := strconv.ParseInt(c.Value, 10, 64); err == nil {
			out[c.Name] = id
		}
	}
	return out
}

func requestRotationPick(r *http.Request, cookieName string) (int64, bool) {
	if r == nil {
		return 0, false
	}
	if picks, ok := r.Context().Value(requestRotationPicksKey).(map[string]int64); ok && picks != nil {
		if id, ok := picks[cookieName]; ok {
			return id, true
		}
	}
	picks := rotationPicksFromRequest(r)
	if id, ok := picks[cookieName]; ok {
		return id, true
	}
	return 0, false
}

func persistRotationPicks(w http.ResponseWriter, r *http.Request, cookiePath string, picks map[string]int64) {
	if w == nil {
		return
	}
	if cookiePath == "" {
		cookiePath = "/"
	}
	if len(picks) == 0 && !hasLegacyRotationCookieHeader(r) {
		return
	}
	merged := rotationPicksFromRequest(r)
	for k, v := range picks {
		merged[k] = v
	}
	merged = trimRotationPicksForCookie(merged, picks)
	encoded, ok := encodeRotationPicksCookie(merged)
	if !ok {
		merged = trimRotationPicksForCookie(merged, nil)
		encoded, ok = encodeRotationPicksCookie(merged)
	}
	if !ok {
		encoded, _ = encodeRotationPicksCookie(map[string]int64{})
	}
	if encoded != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     rotationPicksCookieName,
			Value:    encoded,
			Path:     cookiePath,
			MaxAge:   86400 * 7,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}
	clearLegacyRotationCookies(w, r, cookiePath)
}

func rotationCookiePath() string {
	cookiePath := conf.Conf().BasePath
	if cookiePath == "" {
		return "/"
	}
	return cookiePath
}

func encodeRotationPicksCookie(merged map[string]int64) (string, bool) {
	raw, err := json.Marshal(merged)
	if err != nil {
		return "", false
	}
	encoded := base64.RawURLEncoding.EncodeToString(raw)
	if len(encoded) > maxRotationCookieBytes {
		return "", false
	}
	return encoded, true
}

func decodeRotationPicksCookie(value string, out map[string]int64) {
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		_ = json.Unmarshal([]byte(value), &out)
		return
	}
	_ = json.Unmarshal(raw, &out)
}

func trimRotationPicksForCookie(merged map[string]int64, keep map[string]int64) map[string]int64 {
	for {
		if _, ok := encodeRotationPicksCookie(merged); ok {
			return merged
		}
		removed := false
		for name := range merged {
			if keep != nil {
				if _, mustKeep := keep[name]; mustKeep {
					continue
				}
			}
			if strings.Contains(name, "_p") {
				delete(merged, name)
				removed = true
				break
			}
		}
		if !removed {
			return merged
		}
	}
}

func hasLegacyRotationCookieHeader(r *http.Request) bool {
	if r == nil {
		return false
	}
	for _, c := range r.Cookies() {
		if c.Name == rotationPicksCookieName {
			continue
		}
		if strings.HasPrefix(c.Name, cookiePrefix) {
			return true
		}
	}
	return false
}

func clearLegacyRotationCookies(w http.ResponseWriter, r *http.Request, cookiePath string) {
	if r == nil {
		return
	}
	cleared := 0
	for _, c := range r.Cookies() {
		if cleared >= maxLegacyRotationClearsPerResponse {
			return
		}
		if c.Name == rotationPicksCookieName {
			continue
		}
		if !strings.HasPrefix(c.Name, cookiePrefix) {
			continue
		}
		http.SetCookie(w, &http.Cookie{
			Name:     c.Name,
			Value:    "",
			Path:     cookiePath,
			MaxAge:   -1,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		cleared++
	}
}
