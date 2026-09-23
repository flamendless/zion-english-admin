package cmd

import (
	"net/http"
	"net/url"
	"zion-english/internal/conf"
)

const (
	cookieErrorFlash   = "error_flash"
	cookieSuccessFlash = "success_flash"
)

func setSuccessFlash(w http.ResponseWriter, msg string) {
	setFlashCookie(w, cookieSuccessFlash, msg, false)
}

func setErrorFlash(w http.ResponseWriter, msg string) {
	setFlashCookie(w, cookieErrorFlash, msg, false)
}

func setFlashCookie(w http.ResponseWriter, name, msg string, httpOnly bool) {
	cfg := conf.Conf()
	cookie := &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(msg),
		Path:     cfg.BasePath,
		SameSite: http.SameSiteStrictMode,
	}
	if httpOnly {
		cookie.HttpOnly = true
	}
	if cfg.IsProd() {
		cookie.Secure = true
	}
	http.SetCookie(w, cookie)
}

func readFlashCookie(w http.ResponseWriter, r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil || cookie.Value == "" {
		return ""
	}

	msg, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		msg = cookie.Value
	}

	cfg := conf.Conf()
	clearCookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     cfg.BasePath,
		MaxAge:   -1,
		SameSite: http.SameSiteStrictMode,
	}
	if cfg.IsProd() {
		clearCookie.Secure = true
	}
	http.SetCookie(w, clearCookie)

	return msg
}
