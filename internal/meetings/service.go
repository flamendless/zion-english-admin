package meetings

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"zion-english/internal/database/queries"
	"zion-english/internal/utils"
)

type Config struct {
	Secret         string
	DefaultService string
	Zoom           ZoomConfig
}

type Service struct {
	q          *queries.Queries
	providers  map[string]Provider
	secret     string
	defaultSvc string
	zoom       *ZoomProvider
}

func New(q *queries.Queries, cfg Config) *Service {
	zoomProvider := NewZoomProvider(cfg.Zoom)
	providers := map[string]Provider{
		ServiceZoom: zoomProvider,
	}
	defaultSvc := cfg.DefaultService
	if defaultSvc == "" {
		defaultSvc = ServiceZoom
	}
	return &Service{
		q:          q,
		providers:  providers,
		secret:     cfg.Secret,
		defaultSvc: defaultSvc,
		zoom:       zoomProvider,
	}
}

func (s *Service) ZoomProvider() *ZoomProvider {
	return s.zoom
}

func (s *Service) IsZoomConfigured() bool {
	return s.zoom != nil && s.zoom.IsConfigured()
}

func (s *Service) IsTeacherConnected(ctx context.Context, teacherID int64, service string) (bool, error) {
	if service == "" {
		service = s.defaultSvc
	}
	count, err := s.q.HasTeacherMeetingAccount(ctx, queries.HasTeacherMeetingAccountParams{
		TeacherID: teacherID,
		Service:   service,
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Service) SaveOAuthAccount(ctx context.Context, teacherID int64, account AccountRef, expiresAt time.Time) error {
	accessEnc, err := EncryptToken(s.secret, account.AccessToken)
	if err != nil {
		return err
	}
	refreshEnc, err := EncryptToken(s.secret, account.RefreshToken)
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
		ResourceID:     "",
		AccessToken:    accessEnc,
		RefreshToken:   refreshEnc,
		TokenExpiresAt: expires,
	})
}

func (s *Service) DisconnectTeacher(ctx context.Context, teacherID int64, service string) error {
	if service == "" {
		service = s.defaultSvc
	}
	row, err := s.q.GetTeacherMeetingAccount(ctx, queries.GetTeacherMeetingAccountParams{
		TeacherID: teacherID,
		Service:   service,
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
	if s.zoom != nil {
		_ = s.zoom.RevokeToken(ctx, account.AccessToken)
	}
	return s.q.DeleteTeacherMeetingAccount(ctx, queries.DeleteTeacherMeetingAccountParams{
		TeacherID: teacherID,
		Service:   service,
	})
}

type ScheduledClassMeetingInput struct {
	ScheduleID      int64
	TeacherID       int64
	StudentName     string
	ScheduledDate   string
	StartTime       string
	DurationMinutes int64
}

type ClassMeetingView struct {
	RoomURL  string
	Passcode string
	Service  string
}

func (s *Service) SyncRoomForSchedule(ctx context.Context, input ScheduledClassMeetingInput) error {
	if !SupportsAutoRoom(input.DurationMinutes) {
		return s.removeActiveRoomIfAny(ctx, input.ScheduleID, input.TeacherID)
	}
	provider, account, err := s.providerAccount(ctx, input.TeacherID, s.defaultSvc)
	if err != nil {
		return err
	}
	existing, err := s.q.GetActiveClassMeetingRoomByClassID(ctx, input.ScheduleID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	topic := fmt.Sprintf("Zion English - %s", input.StudentName)
	if err == sql.ErrNoRows {
		room, err := provider.CreateRoom(ctx, account, CreateRoomRequest{
			Topic:           topic,
			ScheduledDate:   input.ScheduledDate,
			StartTime:       input.StartTime,
			DurationMinutes: input.DurationMinutes,
		})
		if err != nil {
			return err
		}
		_, err = s.q.InsertClassMeetingRoom(ctx, queries.InsertClassMeetingRoomParams{
			ClassID:      input.ScheduleID,
			Service:      room.Service,
			RoomID:       room.RoomID,
			RoomUrl:      room.RoomURL,
			RoomPasscode: utils.NullIfEmptyString(room.Passcode),
		})
		return err
	}
	room, err := provider.UpdateRoom(ctx, account, existing.RoomID, UpdateRoomRequest{
		ScheduledDate:   input.ScheduledDate,
		StartTime:       input.StartTime,
		DurationMinutes: input.DurationMinutes,
	})
	if err != nil {
		return err
	}
	roomURL := room.RoomURL
	if roomURL == "" {
		roomURL = existing.RoomUrl
	}
	passcode := room.Passcode
	if passcode == "" {
		passcode = utils.InterfaceToString(existing.RoomPasscode)
	}
	return s.q.UpdateClassMeetingRoom(ctx, queries.UpdateClassMeetingRoomParams{
		RoomUrl:      roomURL,
		RoomPasscode: utils.NullIfEmptyString(passcode),
		ID:           existing.ID,
	})
}

func (s *Service) DeleteRoomForSchedule(ctx context.Context, scheduleID, teacherID int64) error {
	return s.removeActiveRoomIfAny(ctx, scheduleID, teacherID)
}

func (s *Service) removeActiveRoomIfAny(ctx context.Context, scheduleID, teacherID int64) error {
	existing, err := s.q.GetActiveClassMeetingRoomByClassID(ctx, scheduleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	provider, account, err := s.providerAccount(ctx, teacherID, existing.Service)
	if err == nil {
		_ = provider.DeleteRoom(ctx, account, existing.RoomID)
	}
	return s.q.SoftDeleteClassMeetingRoom(ctx, existing.ID)
}

func (s *Service) GetRoomsByClassIDs(ctx context.Context, classIDs []int64) (map[int64]ClassMeetingView, error) {
	result := make(map[int64]ClassMeetingView)
	if len(classIDs) == 0 {
		return result, nil
	}
	rows, err := s.q.GetActiveClassMeetingRoomsByClassIDs(ctx, classIDs)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.ClassID] = ClassMeetingView{
			RoomURL:  row.RoomUrl,
			Passcode: utils.InterfaceToString(row.RoomPasscode),
			Service:  row.Service,
		}
	}
	return result, nil
}

func (s *Service) providerAccount(ctx context.Context, teacherID int64, service string) (Provider, AccountRef, error) {
	provider, ok := s.providers[service]
	if !ok {
		return nil, AccountRef{}, ErrProviderNotFound
	}
	row, err := s.q.GetTeacherMeetingAccount(ctx, queries.GetTeacherMeetingAccountParams{
		TeacherID: teacherID,
		Service:   service,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, AccountRef{}, ErrZoomNotConnected
		}
		return nil, AccountRef{}, err
	}
	account, err := s.decryptAccount(row)
	if err != nil {
		return nil, AccountRef{}, err
	}
	account.TeacherID = teacherID
	if s.zoom != nil {
		account, err = s.refreshAccountIfNeeded(ctx, row, account)
		if err != nil {
			return nil, AccountRef{}, err
		}
	}
	return provider, account, nil
}

func (s *Service) decryptAccount(row queries.GetTeacherMeetingAccountRow) (AccountRef, error) {
	access, err := DecryptToken(s.secret, row.AccessToken)
	if err != nil {
		return AccountRef{}, err
	}
	refresh, err := DecryptToken(s.secret, row.RefreshToken)
	if err != nil {
		return AccountRef{}, err
	}
	return AccountRef{
		TeacherID:      row.TeacherID,
		Service:        row.Service,
		ExternalUserID: row.ExternalUserID,
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
	access, refresh, newExpiry, err := s.zoom.RefreshAccessToken(ctx, account.RefreshToken)
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
