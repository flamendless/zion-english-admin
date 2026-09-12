package calendar

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"zion-english/internal/constants"
)

const (
	googleCalendarScope  = calendar.CalendarScope
	googleUserEmailScope = "https://www.googleapis.com/auth/userinfo.email"
)

func normalizeGoogleConfig(cfg GoogleConfig) GoogleConfig {
	return GoogleConfig{
		ClientID:     strings.TrimSpace(cfg.ClientID),
		ClientSecret: strings.TrimSpace(cfg.ClientSecret),
		RedirectURI:  strings.TrimSpace(cfg.RedirectURI),
	}
}

type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type GoogleProvider struct {
	cfg    GoogleConfig
	client *http.Client
}

func NewGoogleProvider(cfg GoogleConfig) *GoogleProvider {
	return &GoogleProvider{
		cfg: normalizeGoogleConfig(cfg),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *GoogleProvider) IsConfigured() bool {
	return p.cfg.ClientID != "" && p.cfg.ClientSecret != "" && p.cfg.RedirectURI != ""
}

func (p *GoogleProvider) Service() string {
	return ServiceGoogleCalendar
}

func (p *GoogleProvider) RedirectURI() string {
	return p.cfg.RedirectURI
}

func (p *GoogleProvider) oauthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     p.cfg.ClientID,
		ClientSecret: p.cfg.ClientSecret,
		RedirectURL:  p.cfg.RedirectURI,
		Scopes:       []string{googleCalendarScope, googleUserEmailScope},
		Endpoint:     google.Endpoint,
	}
}

func (p *GoogleProvider) AuthorizeURL(state string) (string, error) {
	if !p.IsConfigured() {
		return "", ErrGoogleCalendarNotConfigured
	}
	return p.oauthConfig().AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce), nil
}

func (p *GoogleProvider) ExchangeCode(ctx context.Context, code string) (AccountRef, time.Time, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return AccountRef{}, time.Time{}, errors.New("authorization code is required")
	}
	ctx = p.oauthContext(ctx)
	token, err := p.oauthConfig().Exchange(ctx, code)
	if err != nil {
		return AccountRef{}, time.Time{}, fmt.Errorf("token exchange: %w", err)
	}
	return AccountRef{
		Service:      ServiceGoogleCalendar,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}, token.Expiry, nil
}

func (p *GoogleProvider) oauthContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, oauth2.HTTPClient, p.client)
}

func (p *GoogleProvider) RefreshAccessToken(ctx context.Context, refreshToken string) (string, string, time.Time, error) {
	if refreshToken == "" {
		return "", "", time.Time{}, fmt.Errorf("refresh token is required")
	}
	ctx = p.oauthContext(ctx)
	token, err := p.oauthConfig().TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
		Expiry:       time.Now().Add(-time.Hour),
	}).Token()
	if err != nil {
		return "", "", time.Time{}, err
	}
	refresh := token.RefreshToken
	if refresh == "" {
		refresh = refreshToken
	}
	return token.AccessToken, refresh, token.Expiry, nil
}

func (p *GoogleProvider) RevokeToken(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/revoke?token="+url.QueryEscape(token), nil)
	if err != nil {
		return err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return nil
}

func (p *GoogleProvider) EnsureDedicatedCalendar(ctx context.Context, account AccountRef) (string, error) {
	if account.ResourceID != "" {
		return account.ResourceID, nil
	}
	svc, err := p.calendarService(ctx, account)
	if err != nil {
		return "", err
	}
	list, err := svc.CalendarList.List().Do()
	if err != nil {
		return "", err
	}
	for _, item := range list.Items {
		if strings.EqualFold(strings.TrimSpace(item.Summary), DedicatedCalendarName) {
			return item.Id, nil
		}
	}
	created, err := svc.Calendars.Insert(&calendar.Calendar{
		Summary:     DedicatedCalendarName,
		TimeZone:    constants.TimezoneNamePHT,
		Description: "Scheduled classes from Zion English Admin",
	}).Do()
	if err != nil {
		return "", err
	}
	if created.Id == "" {
		return "", fmt.Errorf("google calendar create returned empty id")
	}
	return created.Id, nil
}

func (p *GoogleProvider) CreateEvent(ctx context.Context, account AccountRef, req CreateEventRequest) (Event, error) {
	start, end, err := eventDateTimes(req.ScheduledDate, req.StartTime, req.DurationMinutes)
	if err != nil {
		return Event{}, err
	}
	svc, err := p.calendarService(ctx, account)
	if err != nil {
		return Event{}, err
	}
	created, err := svc.Events.Insert(account.ResourceID, calendarEvent(req.Summary, req.Description, start, end)).Do()
	if err != nil {
		return Event{}, err
	}
	return Event{
		Service:  ServiceGoogleCalendar,
		EventID:  created.Id,
		EventURL: created.HtmlLink,
	}, nil
}

func (p *GoogleProvider) UpdateEvent(ctx context.Context, account AccountRef, eventID string, req UpdateEventRequest) (Event, error) {
	start, end, err := eventDateTimes(req.ScheduledDate, req.StartTime, req.DurationMinutes)
	if err != nil {
		return Event{}, err
	}
	svc, err := p.calendarService(ctx, account)
	if err != nil {
		return Event{}, err
	}
	updated, err := svc.Events.Patch(account.ResourceID, eventID, calendarEvent(req.Summary, req.Description, start, end)).Do()
	if err != nil {
		return Event{}, err
	}
	event := Event{
		Service: ServiceGoogleCalendar,
		EventID: eventID,
	}
	if updated != nil && updated.HtmlLink != "" {
		event.EventURL = updated.HtmlLink
	}
	return event, nil
}

func (p *GoogleProvider) DeleteEvent(ctx context.Context, account AccountRef, eventID string) error {
	svc, err := p.calendarService(ctx, account)
	if err != nil {
		return err
	}
	return svc.Events.Delete(account.ResourceID, eventID).Do()
}

func (p *GoogleProvider) calendarService(ctx context.Context, account AccountRef) (*calendar.Service, error) {
	ctx = p.oauthContext(ctx)
	return calendar.NewService(ctx, option.WithTokenSource(p.tokenSource(ctx, account)))
}

func (p *GoogleProvider) tokenSource(ctx context.Context, account AccountRef) oauth2.TokenSource {
	ctx = p.oauthContext(ctx)
	return p.oauthConfig().TokenSource(ctx, &oauth2.Token{
		AccessToken:  account.AccessToken,
		RefreshToken: account.RefreshToken,
		TokenType:    "Bearer",
	})
}

func GoogleOAuthRetrieveErrorDetails(err error) (status int, body string) {
	var retrieveErr *oauth2.RetrieveError
	if !errors.As(err, &retrieveErr) {
		return 0, ""
	}
	if retrieveErr.Response != nil {
		status = retrieveErr.Response.StatusCode
	}
	return status, strings.TrimSpace(string(retrieveErr.Body))
}

func calendarEvent(summary, description, start, end string) *calendar.Event {
	return &calendar.Event{
		Summary:     summary,
		Description: description,
		Start: &calendar.EventDateTime{
			DateTime: start,
			TimeZone: constants.TimezoneNamePHT,
		},
		End: &calendar.EventDateTime{
			DateTime: end,
			TimeZone: constants.TimezoneNamePHT,
		},
	}
}

func eventDateTimes(scheduledDate, startTime string, durationMinutes int64) (string, string, error) {
	scheduledDate = strings.TrimSpace(scheduledDate)
	startTime = strings.TrimSpace(startTime)
	if scheduledDate == "" || startTime == "" {
		return "", "", ErrStartTimeRequired
	}
	if durationMinutes <= 0 {
		return "", "", fmt.Errorf("duration must be positive")
	}
	start, err := time.ParseInLocation(constants.DateLayout+" "+constants.TimeHMLayout, scheduledDate+" "+startTime, constants.LocationPHT)
	if err != nil {
		return "", "", err
	}
	end := start.Add(time.Duration(durationMinutes) * time.Minute)
	return start.Format(time.RFC3339), end.Format(time.RFC3339), nil
}
