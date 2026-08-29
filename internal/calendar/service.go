package calendar

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"zion-english/internal/database/queries"
	"zion-english/internal/meetings"
	"zion-english/internal/utils"
)

type Config struct {
	Secret string
	Google GoogleConfig
}

type Service struct {
	q        *queries.Queries
	provider *GoogleProvider
	secret   string
}

func New(q *queries.Queries, cfg Config) *Service {
	return &Service{
		q:        q,
		provider: NewGoogleProvider(cfg.Google),
		secret:   cfg.Secret,
	}
}

func (s *Service) GoogleProvider() *GoogleProvider {
	return s.provider
}

func (s *Service) IsConfigured() bool {
	return s.provider != nil && s.provider.IsConfigured()
}

func (s *Service) IsTeacherConnected(ctx context.Context, teacherID int64) (bool, error) {
	count, err := s.q.HasTeacherMeetingAccount(ctx, queries.HasTeacherMeetingAccountParams{
		TeacherID: teacherID,
		Service:   ServiceGoogleCalendar,
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Service) SaveOAuthAccount(ctx context.Context, teacherID int64, account AccountRef, expiresAt time.Time) error {
	accessEnc, err := meetings.EncryptToken(s.secret, account.AccessToken)
	if err != nil {
		return err
	}
	refreshEnc, err := meetings.EncryptToken(s.secret, account.RefreshToken)
	if err != nil {
		return err
	}
	var expires any
	if !expiresAt.IsZero() {
		expires = expiresAt.UTC().Format(time.RFC3339)
	}
	return s.q.UpsertTeacherMeetingAccount(ctx, queries.UpsertTeacherMeetingAccountParams{
		TeacherID:      teacherID,
		Service:        account.Service,
		ExternalUserID: account.ExternalUserID,
		ResourceID:     account.ResourceID,
		AccessToken:    accessEnc,
		RefreshToken:   refreshEnc,
		TokenExpiresAt: expires,
	})
}

func (s *Service) DisconnectTeacher(ctx context.Context, teacherID int64) error {
	row, err := s.q.GetTeacherMeetingAccount(ctx, queries.GetTeacherMeetingAccountParams{
		TeacherID: teacherID,
		Service:   ServiceGoogleCalendar,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	account, err := s.decryptAccount(row)
	if err != nil {
		return err
	}
	if s.provider != nil {
		_ = s.provider.RevokeToken(ctx, account.AccessToken)
	}
	return s.q.DeleteTeacherMeetingAccount(ctx, queries.DeleteTeacherMeetingAccountParams{
		TeacherID: teacherID,
		Service:   ServiceGoogleCalendar,
	})
}

type ScheduledClassCalendarInput struct {
	ScheduleID      int64
	TeacherID       int64
	StudentName     string
	ScheduledDate   string
	StartTime       string
	DurationMinutes int64
	Rate            float64
	Currency        string
}

type CalendarEventView struct {
	EventURL string
	Service  string
}

func (s *Service) SyncEventForSchedule(ctx context.Context, input ScheduledClassCalendarInput) error {
	if strings.TrimSpace(input.StartTime) == "" {
		return nil
	}
	account, err := s.providerAccount(ctx, input.TeacherID)
	if err != nil {
		return err
	}
	existing, err := s.q.GetActiveClassCalendarEventByClassID(ctx, input.ScheduleID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	summary := fmt.Sprintf("Zion English - %s", input.StudentName)
	description := buildEventDescription(input)
	createReq := CreateEventRequest{
		Summary:         summary,
		Description:     description,
		ScheduledDate:   input.ScheduledDate,
		StartTime:       input.StartTime,
		DurationMinutes: input.DurationMinutes,
	}
	updateReq := UpdateEventRequest{
		Summary:         summary,
		Description:     description,
		ScheduledDate:   input.ScheduledDate,
		StartTime:       input.StartTime,
		DurationMinutes: input.DurationMinutes,
	}
	if err == sql.ErrNoRows {
		event, err := s.provider.CreateEvent(ctx, account, createReq)
		if err != nil {
			return err
		}
		_, err = s.q.InsertClassCalendarEvent(ctx, queries.InsertClassCalendarEventParams{
			ClassID:  input.ScheduleID,
			Service:  event.Service,
			EventID:  event.EventID,
			EventUrl: event.EventURL,
		})
		return err
	}
	event, err := s.provider.UpdateEvent(ctx, account, existing.EventID, updateReq)
	if err != nil {
		return err
	}
	eventURL := event.EventURL
	if eventURL == "" {
		eventURL = existing.EventUrl
	}
	return s.q.UpdateClassCalendarEvent(ctx, queries.UpdateClassCalendarEventParams{
		EventUrl: eventURL,
		ID:       existing.ID,
	})
}

func (s *Service) DeleteEventForSchedule(ctx context.Context, scheduleID, teacherID int64) error {
	existing, err := s.q.GetActiveClassCalendarEventByClassID(ctx, scheduleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	account, err := s.providerAccount(ctx, teacherID)
	if err == nil {
		_ = s.provider.DeleteEvent(ctx, account, existing.EventID)
	}
	return s.q.SoftDeleteClassCalendarEvent(ctx, existing.ID)
}

func (s *Service) GetEventsByClassIDs(ctx context.Context, classIDs []int64) (map[int64]CalendarEventView, error) {
	result := make(map[int64]CalendarEventView)
	if len(classIDs) == 0 {
		return result, nil
	}
	rows, err := s.q.GetActiveClassCalendarEventsByClassIDs(ctx, classIDs)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.ClassID] = CalendarEventView{
			EventURL: row.EventUrl,
			Service:  row.Service,
		}
	}
	return result, nil
}

func (s *Service) ConnectAccount(ctx context.Context, teacherID int64, account AccountRef, expiresAt time.Time) error {
	account.TeacherID = teacherID
	calendarID, err := s.provider.EnsureDedicatedCalendar(ctx, account)
	if err != nil {
		return err
	}
	account.ResourceID = calendarID
	return s.SaveOAuthAccount(ctx, teacherID, account, expiresAt)
}

func (s *Service) providerAccount(ctx context.Context, teacherID int64) (AccountRef, error) {
	row, err := s.q.GetTeacherMeetingAccount(ctx, queries.GetTeacherMeetingAccountParams{
		TeacherID: teacherID,
		Service:   ServiceGoogleCalendar,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return AccountRef{}, ErrGoogleCalendarNotConnected
		}
		return AccountRef{}, err
	}
	account, err := s.decryptAccount(row)
	if err != nil {
		return AccountRef{}, err
	}
	account.TeacherID = teacherID
	account, err = s.refreshAccountIfNeeded(ctx, row, account)
	if err != nil {
		return AccountRef{}, err
	}
	if account.ResourceID == "" {
		calendarID, err := s.provider.EnsureDedicatedCalendar(ctx, account)
		if err != nil {
			return AccountRef{}, err
		}
		account.ResourceID = calendarID
		if err := s.SaveOAuthAccount(ctx, teacherID, account, time.Time{}); err != nil {
			return AccountRef{}, err
		}
	}
	return account, nil
}

func (s *Service) decryptAccount(row queries.GetTeacherMeetingAccountRow) (AccountRef, error) {
	access, err := meetings.DecryptToken(s.secret, row.AccessToken)
	if err != nil {
		return AccountRef{}, err
	}
	refresh, err := meetings.DecryptToken(s.secret, row.RefreshToken)
	if err != nil {
		return AccountRef{}, err
	}
	return AccountRef{
		TeacherID:      row.TeacherID,
		Service:        row.Service,
		ExternalUserID: row.ExternalUserID,
		ResourceID:     row.ResourceID,
		AccessToken:    access,
		RefreshToken:   refresh,
	}, nil
}

func (s *Service) refreshAccountIfNeeded(ctx context.Context, row queries.GetTeacherMeetingAccountRow, account AccountRef) (AccountRef, error) {
	expiresRaw := utils.InterfaceToString(row.TokenExpiresAt)
	if expiresRaw == "" {
		return account, nil
	}
	expiresAt, err := time.Parse(time.RFC3339, expiresRaw)
	if err != nil {
		return account, nil
	}
	if time.Now().UTC().Add(2 * time.Minute).Before(expiresAt) {
		return account, nil
	}
	access, refresh, newExpiry, err := s.provider.RefreshAccessToken(ctx, account.RefreshToken)
	if err != nil {
		return AccountRef{}, err
	}
	account.AccessToken = access
	account.RefreshToken = refresh
	if err := s.SaveOAuthAccount(ctx, row.TeacherID, account, newExpiry); err != nil {
		return AccountRef{}, err
	}
	return account, nil
}

func buildEventDescription(input ScheduledClassCalendarInput) string {
	endTime := input.StartTime
	if input.DurationMinutes > 0 {
		if start, err := time.ParseInLocation("15:04", strings.TrimSpace(input.StartTime), time.UTC); err == nil {
			end := start.Add(time.Duration(input.DurationMinutes) * time.Minute)
			endTime = end.Format("15:04")
		}
	}
	lines := []string{
		fmt.Sprintf("Student: %s", input.StudentName),
		fmt.Sprintf("Date: %s", input.ScheduledDate),
		fmt.Sprintf("Time: %s - %s (PHT)", input.StartTime, endTime),
	}
	if input.Rate > 0 && input.Currency != "" {
		lines = append(lines, fmt.Sprintf("Rate: %.2f %s", input.Rate, input.Currency))
	}
	return strings.Join(lines, "\n")
}
