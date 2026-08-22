package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"
	"time"
	"zion-english/internal/conf"
)

const csrfCookieName = "csrf_token"
const csrfHeaderName = "X-CSRF-Token"
const csrfFormField = "csrf_token"
const csrfMultipartMaxBytes = 5 << 20

func csrfCookiePath(cfg *conf.Config) string {
	path := cfg.BasePath
	if path == "" {
		return "/"
	}
	return path
}

func EnsureCSRFToken(w http.ResponseWriter, r *http.Request, cfg *conf.Config) {
	if _, err := r.Cookie(csrfCookieName); err == nil {
		return
	}

	token, err := newCSRFToken()
	if err != nil {
		return
	}

	cookie := &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     csrfCookiePath(cfg),
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(24 * time.Hour),
	}
	if cfg.IsProd() {
		cookie.Secure = true
	}
	http.SetCookie(w, cookie)
}

func newCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func ValidateCSRF(r *http.Request) bool {
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	}

	cookie, err := r.Cookie(csrfCookieName)
	if err != nil || cookie.Value == "" {
		return false
	}

	token := csrfTokenFromRequest(r)
	if token == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(token)) == 1
}

func csrfTokenFromRequest(r *http.Request) string {
	if token := r.Header.Get(csrfHeaderName); token != "" {
		return token
	}

	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(csrfMultipartMaxBytes); err != nil {
			return ""
		}
		return r.FormValue(csrfFormField)
	}

	if err := r.ParseForm(); err != nil {
		return ""
	}
	return r.FormValue(csrfFormField)
}

func CSRFMiddleware(cfg *conf.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		EnsureCSRFToken(w, r, cfg)

		if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions {
			if !ValidateCSRF(r) {
				http.Error(w, "Invalid or missing CSRF token", http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
