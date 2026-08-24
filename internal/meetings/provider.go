package meetings

import "context"

type AccountRef struct {
	TeacherID      int64
	Service        string
	ExternalUserID string
	AccessToken    string
	RefreshToken   string
}

type CreateRoomRequest struct {
	Topic           string
	ScheduledDate   string
	StartTime       string
	DurationMinutes int64
}

type UpdateRoomRequest struct {
	ScheduledDate   string
	StartTime       string
	DurationMinutes int64
}

type Room struct {
	Service  string
	RoomID   string
	RoomURL  string
	Passcode string
}

type Provider interface {
	Service() string
	CreateRoom(ctx context.Context, account AccountRef, req CreateRoomRequest) (Room, error)
	UpdateRoom(ctx context.Context, account AccountRef, roomID string, req UpdateRoomRequest) (Room, error)
	DeleteRoom(ctx context.Context, account AccountRef, roomID string) error
}
