package ads

import (
	"context"
	"net/http"
	"strings"

	"zion-english/internal/conf"
)

const rotationPicksCookieName = "zion_ad_picks"

// maxLegacyRotationClearsPerPartialResponse keeps HTMX fragment responses under proxy Set-Cookie limits.
const maxLegacyRotationClearsPerPartialResponse = 24

type requestRotationPicksKeyType struct{}

var requestRotationPicksKey = requestRotationPicksKeyType{}

func withRequestRotationPicks(ctx context.Context, _ *http.Request) context.Context {
	return context.WithValue(ctx, requestRotationPicksKey, map[string]int64{})
}

func rotationCookiePath() string {
	cookiePath := conf.Conf().BasePath
	if cookiePath == "" {
		return "/"
	}
	return cookiePath
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

func needsRotationCookiePurge(r *http.Request) bool {
	if r == nil {
		return false
	}
	if hasLegacyRotationCookieHeader(r) {
		return true
	}
	if _, err := r.Cookie(rotationPicksCookieName); err == nil {
		return true
	}
	return false
}

func legacyRotationCookieCount(r *http.Request) int {
	if r == nil {
		return 0
	}
	seen := make(map[string]struct{})
	n := 0
	for _, c := range r.Cookies() {
		if c.Name == rotationPicksCookieName {
			continue
		}
		if !strings.HasPrefix(c.Name, cookiePrefix) {
			continue
		}
		if _, ok := seen[c.Name]; ok {
			continue
		}
		seen[c.Name] = struct{}{}
		n++
	}
	return n
}

// PurgeLegacyRotationCookies drops bloated rotation state and clears every legacy zion_ad_* cookie.
func PurgeLegacyRotationCookies(w http.ResponseWriter, r *http.Request) {
	if w == nil || !needsRotationCookiePurge(r) {
		return
	}
	if r != nil && isPublicCookieRecoveryPath(r.URL.Path) &&
		legacyRotationCookieCount(r) > maxLegacyRotationClearsPerPartialResponse {
		w.Header().Add("Clear-Site-Data", `"cookies"`)
		return
	}
	cookiePath := rotationCookiePath()
	clearRotationPicksCookie(w, cookiePath)
	clearLegacyRotationCookies(w, r, cookiePath, legacyRotationClearLimit(r))
}

func legacyRotationClearLimit(r *http.Request) int {
	if r != nil && strings.Contains(r.URL.Path, "/partials/") {
		return maxLegacyRotationClearsPerPartialResponse
	}
	return 0
}

func clearLegacyRotationCookies(w http.ResponseWriter, r *http.Request, cookiePath string, maxClears int) {
	if r == nil {
		return
	}
	paths := rotationCookieClearPaths(cookiePath)
	if maxClears > 0 {
		if cookiePath == "" {
			cookiePath = "/"
		}
		paths = []string{cookiePath}
	}
	seen := make(map[string]struct{})
	cleared := 0
	for _, c := range r.Cookies() {
		if c.Name == rotationPicksCookieName {
			continue
		}
		if !strings.HasPrefix(c.Name, cookiePrefix) {
			continue
		}
		if _, ok := seen[c.Name]; ok {
			continue
		}
		seen[c.Name] = struct{}{}
		for _, path := range paths {
			if maxClears > 0 && cleared >= maxClears {
				return
			}
			http.SetCookie(w, &http.Cookie{
				Name:     c.Name,
				Value:    "",
				Path:     path,
				MaxAge:   -1,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
			if maxClears > 0 {
				cleared++
			}
		}
	}
}

func clearRotationPicksCookie(w http.ResponseWriter, cookiePath string) {
	if w == nil {
		return
	}
	if cookiePath == "" {
		cookiePath = "/"
	}
	cookie := &http.Cookie{
		Name:     rotationPicksCookieName,
		Value:    "",
		Path:     cookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	if conf.Conf().IsProd() {
		cookie.Secure = true
	}
	http.SetCookie(w, cookie)
}

func rotationCookieClearPaths(cookiePath string) []string {
	if cookiePath == "" || cookiePath == "/" {
		return []string{"/"}
	}
	return []string{cookiePath, "/"}
}
