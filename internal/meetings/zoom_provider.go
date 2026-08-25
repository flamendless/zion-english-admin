package meetings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"zion-english/internal/constants"
)

const (
	zoomOAuthAuthorizeURL = "https://zoom.us/oauth/authorize"
	zoomOAuthTokenURL     = "https://zoom.us/oauth/token"
	zoomAPIBaseURL        = "https://api.zoom.us/v2"
)

type ZoomConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type ZoomProvider struct {
	cfg    ZoomConfig
	client *http.Client
}

func NewZoomProvider(cfg ZoomConfig) *ZoomProvider {
	return &ZoomProvider{
		cfg: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *ZoomProvider) IsConfigured() bool {
	return p.cfg.ClientID != "" && p.cfg.ClientSecret != "" && p.cfg.RedirectURI != ""
}

func (p *ZoomProvider) Service() string {
	return ServiceZoom
}

// AuthorizeURL builds the Zoom OAuth consent URL. Required scopes must be enabled on the
// Marketplace app (not passed here — Zoom rejects authorize URLs with scope query params):
// user:read:user, meeting:write:meeting, meeting:update:meeting, meeting:delete:meeting.
func (p *ZoomProvider) AuthorizeURL(state string) (string, error) {
	if p.cfg.ClientID == "" || p.cfg.RedirectURI == "" {
		return "", ErrZoomNotConfigured
	}
	values := url.Values{}
	values.Set("response_type", "code")
	values.Set("client_id", p.cfg.ClientID)
	values.Set("redirect_uri", p.cfg.RedirectURI)
	values.Set("state", state)
	return zoomOAuthAuthorizeURL + "?" + values.Encode(), nil
}

type zoomTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type zoomUserResponse struct {
	ID string `json:"id"`
}

type zoomMeetingResponse struct {
	ID       int64  `json:"id"`
	JoinURL  string `json:"join_url"`
	Password string `json:"password"`
}

func (p *ZoomProvider) ExchangeCode(ctx context.Context, code string) (AccountRef, time.Time, error) {
	token, expiresAt, err := p.requestToken(ctx, url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {p.cfg.RedirectURI},
	})
	if err != nil {
		return AccountRef{}, time.Time{}, err
	}
	user, err := p.getCurrentUser(ctx, token.AccessToken)
	if err != nil {
		return AccountRef{}, time.Time{}, err
	}
	return AccountRef{
		Service:        ServiceZoom,
		ExternalUserID: user.ID,
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
	}, expiresAt, nil
}

func (p *ZoomProvider) RefreshAccessToken(ctx context.Context, refreshToken string) (string, string, time.Time, error) {
	token, expiresAt, err := p.requestToken(ctx, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	})
	if err != nil {
		return "", "", time.Time{}, err
	}
	refresh := token.RefreshToken
	if refresh == "" {
		refresh = refreshToken
	}
	return token.AccessToken, refresh, expiresAt, nil
}

func (p *ZoomProvider) RevokeToken(ctx context.Context, token string) error {
	if token == "" || p.cfg.ClientID == "" || p.cfg.ClientSecret == "" {
		return nil
	}
	body := url.Values{"token": {token}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://zoom.us/oauth/revoke", strings.NewReader(body.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(p.cfg.ClientID, p.cfg.ClientSecret)
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return nil
}

func (p *ZoomProvider) CreateRoom(ctx context.Context, account AccountRef, req CreateRoomRequest) (Room, error) {
	startTime, err := zoomMeetingsStartTime(req.ScheduledDate, req.StartTime)
	if err != nil {
		return Room{}, err
	}
	payload := map[string]interface{}{
		"topic":      req.Topic,
		"type":       2,
		"start_time": startTime,
		"duration":   req.DurationMinutes,
		"timezone":   constants.TimezoneNamePHT,
	}
	var resp zoomMeetingResponse
	if err := p.apiRequest(ctx, account.AccessToken, http.MethodPost, "/users/me/meetings", payload, &resp); err != nil {
		return Room{}, err
	}
	return Room{
		Service:  ServiceZoom,
		RoomID:   fmt.Sprintf("%d", resp.ID),
		RoomURL:  resp.JoinURL,
		Passcode: resp.Password,
	}, nil
}

func (p *ZoomProvider) UpdateRoom(ctx context.Context, account AccountRef, roomID string, req UpdateRoomRequest) (Room, error) {
	startTime, err := zoomMeetingsStartTime(req.ScheduledDate, req.StartTime)
	if err != nil {
		return Room{}, err
	}
	payload := map[string]interface{}{
		"start_time": startTime,
		"duration":   req.DurationMinutes,
		"timezone":   constants.TimezoneNamePHT,
	}
	path := "/meetings/" + url.PathEscape(roomID)
	var resp zoomMeetingResponse
	if err := p.apiRequest(ctx, account.AccessToken, http.MethodPatch, path, payload, &resp); err != nil {
		return Room{}, err
	}
	room := Room{
		Service: ServiceZoom,
		RoomID:  roomID,
	}
	if resp.JoinURL != "" {
		room.RoomURL = resp.JoinURL
	}
	if resp.Password != "" {
		room.Passcode = resp.Password
	}
	return room, nil
}

func (p *ZoomProvider) DeleteRoom(ctx context.Context, account AccountRef, roomID string) error {
	path := "/meetings/" + url.PathEscape(roomID)
	return p.apiRequest(ctx, account.AccessToken, http.MethodDelete, path, nil, nil)
}

func (p *ZoomProvider) requestToken(ctx context.Context, body url.Values) (zoomTokenResponse, time.Time, error) {
	if p.cfg.ClientID == "" || p.cfg.ClientSecret == "" {
		return zoomTokenResponse{}, time.Time{}, ErrZoomNotConfigured
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, zoomOAuthTokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return zoomTokenResponse{}, time.Time{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(p.cfg.ClientID, p.cfg.ClientSecret)
	resp, err := p.client.Do(req)
	if err != nil {
		return zoomTokenResponse{}, time.Time{}, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return zoomTokenResponse{}, time.Time{}, err
	}
	if resp.StatusCode >= 400 {
		return zoomTokenResponse{}, time.Time{}, fmt.Errorf("zoom token request failed: %s", strings.TrimSpace(string(raw)))
	}
	var token zoomTokenResponse
	if err := json.Unmarshal(raw, &token); err != nil {
		return zoomTokenResponse{}, time.Time{}, err
	}
	expiresAt := time.Now().UTC().Add(time.Duration(token.ExpiresIn) * time.Second)
	return token, expiresAt, nil
}

func (p *ZoomProvider) getCurrentUser(ctx context.Context, accessToken string) (zoomUserResponse, error) {
	var user zoomUserResponse
	if err := p.apiRequest(ctx, accessToken, http.MethodGet, "/users/me", nil, &user); err != nil {
		return zoomUserResponse{}, err
	}
	return user, nil
}

func (p *ZoomProvider) apiRequest(ctx context.Context, accessToken, method, path string, payload interface{}, out interface{}) error {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, zoomAPIBaseURL+path, body)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("zoom api %s %s failed: %s", method, path, strings.TrimSpace(string(raw)))
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

func zoomMeetingsStartTime(scheduledDate, startTime string) (string, error) {
	scheduledDate = strings.TrimSpace(scheduledDate)
	startTime = strings.TrimSpace(startTime)
	if scheduledDate == "" || startTime == "" {
		return "", fmt.Errorf("scheduled date and start time are required")
	}
	t, err := time.ParseInLocation(constants.DateLayout+" "+constants.TimeHMLayout, scheduledDate+" "+startTime, constants.LocationPHT)
	if err != nil {
		return "", err
	}
	return t.Format("2006-01-02T15:04:05"), nil
}
