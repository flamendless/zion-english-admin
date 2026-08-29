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
	"zion-english/internal/conf"
	"zion-english/internal/constants"
	"zion-english/internal/featureflags"
	"zion-english/internal/logs"
	"zion-english/internal/meetings"

	"go.uber.org/zap"
)

const (
	zoomOAuthStateCookie = "zoom_oauth_state"
	zoomOAuthStateTTL    = 10 * time.Minute
)

func handleZoomConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if meetingSvc == nil || !meetingSvc.IsZoomConfigured() {
		HttpError(w, "Zoom integration is not configured", http.StatusServiceUnavailable)
		return
	}
	if !featureflags.IsEnabled(r.Context(), dbRO, constants.FeatureFlagIntegrationZoom) {
		HttpError(w, "Zoom connections are currently disabled", http.StatusServiceUnavailable)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	cfg := conf.Conf()
	auth.RefreshSessionCookie(w, r, cfg)
	state, err := newZoomOAuthState(cfg.Secret, user.ID)
	if err != nil {
		HttpError(w, "Failed to start Zoom connection", http.StatusInternalServerError)
		return
	}
	setZoomOAuthStateCookie(w, cfg, state)

	url, err := meetingSvc.ZoomProvider().AuthorizeURL(state)
	if err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func handleZoomCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if meetingSvc == nil || !meetingSvc.IsZoomConfigured() {
		HttpError(w, "Zoom integration is not configured", http.StatusServiceUnavailable)
		return
	}

	cfg := conf.Conf()
	ctx := r.Context()

	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		clearZoomOAuthStateCookie(w, cfg)
		HttpRedirect(w, r, "/profile?zoom_error="+urlQueryEscape(errMsg))
		return
	}

	state := r.URL.Query().Get("state")
	cookieState, err := r.Cookie(zoomOAuthStateCookie)
	if err != nil || cookieState.Value == "" || cookieState.Value != state {
		HttpRedirect(w, r, "/profile?zoom_error="+string(constants.IntegrationOAuthErrorInvalidState))
		return
	}
	teacherID, ok := parseZoomOAuthState(cfg.Secret, state)
	if !ok {
		clearZoomOAuthStateCookie(w, cfg)
		HttpRedirect(w, r, "/profile?zoom_error="+string(constants.IntegrationOAuthErrorInvalidState))
		return
	}
	clearZoomOAuthStateCookie(w, cfg)

	code := r.URL.Query().Get("code")
	if code == "" {
		HttpRedirect(w, r, "/profile?zoom_error="+string(constants.IntegrationOAuthErrorMissingCode))
		return
	}

	account, expiresAt, err := meetingSvc.ZoomProvider().ExchangeCode(ctx, code)
	if err != nil {
		logs.Log().Error("zoom oauth exchange failed", zap.Error(err))
		HttpRedirect(w, r, "/profile?zoom_error="+string(constants.IntegrationOAuthErrorExchangeFailed))
		return
	}
	account.Service = meetings.ServiceZoom
	if err := meetingSvc.SaveOAuthAccount(ctx, teacherID, account, expiresAt); err != nil {
		logs.Log().Error("save zoom account failed", zap.Error(err))
		HttpRedirect(w, r, "/profile?zoom_error="+string(constants.IntegrationOAuthErrorSaveFailed))
		return
	}

	insertAuditLogAs(ctx, auth.User{ID: teacherID, Role: auth.RoleTeacher}, "profile", "connected zoom account")
	HttpRedirect(w, r, "/profile?zoom_connected=1")
}

func handleZoomDisconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	user := auth.GetUser(ctx)
	if meetingSvc == nil {
		HttpError(w, "Zoom integration is not configured", http.StatusServiceUnavailable)
		return
	}
	if err := meetingSvc.DisconnectTeacher(ctx, user.ID, meetings.ServiceZoom); err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	insertAuditLogAs(ctx, user, "profile", "disconnected zoom account")
	HttpRedirect(w, r, "/profile?zoom_disconnected=1")
}

func teacherZoomConnected(ctx context.Context, teacherID int64) bool {
	if meetingSvc == nil {
		return false
	}
	ok, err := meetingSvc.IsTeacherConnected(ctx, teacherID, meetings.ServiceZoom)
	return err == nil && ok
}

func newZoomOAuthState(secret string, teacherID int64) (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	expiry := time.Now().UTC().Add(zoomOAuthStateTTL).Unix()
	payload := fmt.Sprintf("%d|%d|%s", teacherID, expiry, hex.EncodeToString(nonce))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return base64.RawURLEncoding.EncodeToString([]byte(payload + "|" + sig)), nil
}

func parseZoomOAuthState(secret, state string) (int64, bool) {
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

func setZoomOAuthStateCookie(w http.ResponseWriter, cfg *conf.Config, state string) {
	cookie := &http.Cookie{
		Name:     zoomOAuthStateCookie,
		Value:    state,
		Path:     oauthCookiePath(cfg),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(zoomOAuthStateTTL.Seconds()),
	}
	if cfg.IsProd() {
		cookie.Secure = true
	}
	http.SetCookie(w, cookie)
}

func clearZoomOAuthStateCookie(w http.ResponseWriter, cfg *conf.Config) {
	cookie := &http.Cookie{
		Name:     zoomOAuthStateCookie,
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

func oauthCookiePath(cfg *conf.Config) string {
	path := cfg.BasePath
	if path == "" {
		return "/"
	}
	return path
}

func urlQueryEscape(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, " ", "+"), "\n", "")
}

func profileZoomStatus(ctx context.Context, teacherID int64) (connected bool, envConfigured bool, connectionsAllowed bool) {
	envConfigured = meetingSvc != nil && meetingSvc.IsZoomConfigured()
	connectionsAllowed = featureflags.IsEnabled(ctx, dbRO, constants.FeatureFlagIntegrationZoom)
	if !envConfigured {
		return false, false, connectionsAllowed
	}
	connected = teacherZoomConnected(ctx, teacherID)
	return connected, envConfigured, connectionsAllowed
}
