package cmd

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"zion-english/internal/auth"
	"zion-english/internal/calendar"
	"zion-english/internal/conf"
	"zion-english/internal/constants"
	"zion-english/internal/featureflags"
	"zion-english/internal/logs"

	"go.uber.org/zap"
)

const (
	googleCalendarOAuthStateCookie = "google_calendar_oauth_state"
	googleCalendarOAuthStateTTL    = 10 * time.Minute
)

func handleGoogleCalendarConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if calendarSvc == nil || !calendarSvc.IsConfigured() {
		HttpError(w, "Google Calendar integration is not configured", http.StatusServiceUnavailable)
		return
	}
	if !featureflags.IsEnabled(r.Context(), dbRO, constants.FeatureFlagIntegrationGoogleCalendar) {
		HttpError(w, "Google Calendar connections are currently disabled", http.StatusServiceUnavailable)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	cfg := conf.Conf()
	auth.RefreshSessionCookie(w, r, cfg)
	state, err := newGoogleCalendarOAuthState(cfg.Secret, user.ID)
	if err != nil {
		HttpError(w, "Failed to start Google Calendar connection", http.StatusInternalServerError)
		return
	}
	setGoogleCalendarOAuthStateCookie(w, cfg, state)

	url, err := calendarSvc.GoogleProvider().AuthorizeURL(state)
	if err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func handleGoogleCalendarCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if calendarSvc == nil || !calendarSvc.IsConfigured() {
		HttpError(w, "Google Calendar integration is not configured", http.StatusServiceUnavailable)
		return
	}

	cfg := conf.Conf()
	ctx := r.Context()

	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		clearGoogleCalendarOAuthStateCookie(w, cfg)
		HttpRedirect(w, r, "/profile?google_calendar_error="+urlQueryEscape(errMsg))
		return
	}

	state := r.URL.Query().Get("state")
	cookieState, err := r.Cookie(googleCalendarOAuthStateCookie)
	if err != nil || cookieState.Value == "" || cookieState.Value != state {
		HttpRedirect(w, r, "/profile?google_calendar_error="+string(constants.IntegrationOAuthErrorInvalidState))
		return
	}
	teacherID, ok := parseGoogleCalendarOAuthState(cfg.Secret, state)
	if !ok {
		clearGoogleCalendarOAuthStateCookie(w, cfg)
		HttpRedirect(w, r, "/profile?google_calendar_error="+string(constants.IntegrationOAuthErrorInvalidState))
		return
	}
	clearGoogleCalendarOAuthStateCookie(w, cfg)

	code := r.URL.Query().Get("code")
	if code == "" {
		HttpRedirect(w, r, "/profile?google_calendar_error="+string(constants.IntegrationOAuthErrorMissingCode))
		return
	}

	account, expiresAt, err := calendarSvc.GoogleProvider().ExchangeCode(ctx, code)
	if err != nil {
		logs.Log().Error("google calendar oauth exchange failed", zap.Error(err))
		HttpRedirect(w, r, "/profile?google_calendar_error="+string(constants.IntegrationOAuthErrorExchangeFailed))
		return
	}
	account.Service = calendar.ServiceGoogleCalendar
	if err := calendarSvc.ConnectAccount(ctx, teacherID, account, expiresAt); err != nil {
		logs.Log().Error("save google calendar account failed", zap.Error(err))
		HttpRedirect(w, r, "/profile?google_calendar_error="+string(constants.IntegrationOAuthErrorSaveFailed))
		return
	}

	insertAuditLogAs(ctx, auth.User{ID: teacherID, Role: auth.RoleTeacher}, "profile", "connected google calendar account")
	HttpRedirect(w, r, "/profile?google_calendar_connected=1")
}

func handleGoogleCalendarDisconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	user := auth.GetUser(ctx)
	if calendarSvc == nil {
		HttpError(w, "Google Calendar integration is not configured", http.StatusServiceUnavailable)
		return
	}
	if err := calendarSvc.DisconnectTeacher(ctx, user.ID); err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	insertAuditLogAs(ctx, user, "profile", "disconnected google calendar account")
	HttpRedirect(w, r, "/profile?google_calendar_disconnected=1")
}

func teacherGoogleCalendarConnected(ctx context.Context, teacherID int64) bool {
	if calendarSvc == nil {
		return false
	}
	ok, err := calendarSvc.IsTeacherConnected(ctx, teacherID)
	return err == nil && ok
}

func newGoogleCalendarOAuthState(secret string, teacherID int64) (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	expiry := time.Now().UTC().Add(googleCalendarOAuthStateTTL).Unix()
	payload := fmt.Sprintf("%d|%d|%s", teacherID, expiry, hex.EncodeToString(nonce))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return base64.RawURLEncoding.EncodeToString([]byte(payload + "|" + sig)), nil
}

func parseGoogleCalendarOAuthState(secret, state string) (int64, bool) {
	raw, err := base64.RawURLEncoding.DecodeString(state)
	if err != nil {
		return 0, false
	}
	parts := strings.Split(string(raw), "|")
	if len(parts) != 4 {
		return 0, false
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	expiry, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || time.Now().UTC().Unix() > expiry {
		return 0, false
	}
	payload := strings.Join(parts[:3], "|")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[3])) {
		return 0, false
	}
	return id, true
}

func setGoogleCalendarOAuthStateCookie(w http.ResponseWriter, cfg *conf.Config, state string) {
	cookie := &http.Cookie{
		Name:     googleCalendarOAuthStateCookie,
		Value:    state,
		Path:     oauthCookiePath(cfg),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(googleCalendarOAuthStateTTL.Seconds()),
	}
	if cfg.IsProd() {
		cookie.Secure = true
	}
	http.SetCookie(w, cookie)
}

func clearGoogleCalendarOAuthStateCookie(w http.ResponseWriter, cfg *conf.Config) {
	cookie := &http.Cookie{
		Name:     googleCalendarOAuthStateCookie,
		Value:    "",
		Path:     oauthCookiePath(cfg),
		HttpOnly: true,
		MaxAge:   -1,
	}
	if cfg.IsProd() {
		cookie.Secure = true
	}
	http.SetCookie(w, cookie)
}

func profileGoogleCalendarStatus(ctx context.Context, teacherID int64) (connected bool, envConfigured bool, connectionsAllowed bool) {
	envConfigured = calendarSvc != nil && calendarSvc.IsConfigured()
	connectionsAllowed = featureflags.IsEnabled(ctx, dbRO, constants.FeatureFlagIntegrationGoogleCalendar)
	if !envConfigured {
		return false, false, connectionsAllowed
	}
	connected = teacherGoogleCalendarConnected(ctx, teacherID)
	return connected, envConfigured, connectionsAllowed
}
