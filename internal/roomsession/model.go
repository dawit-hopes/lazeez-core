package roomsession

import (
	"lazeez-core/internal/common"
	"time"
)

type RoomSession struct {
	common.Base
	SessionKey    string    `json:"session_key" db:"session_key"`
	RoomReference string    `json:"room_reference" db:"room_reference"`
	RoomID        string    `json:"room_id" db:"room_id"`
	BookingID     string    `json:"booking_id" db:"booking_id"`
	BranchID      string    `json:"branch_id" db:"branch_id"`
	ExpiresAt     time.Time `json:"expires_at" db:"expires_at"`
}

func (s *RoomSession) Table() string {
	return "room_sessions"
}

func (s *RoomSession) Columns() []string {
	return []string{
		"id",
		"session_key",
		"room_reference",
		"room_id",
		"booking_id",
		"branch_id",
		"expires_at",
		"deleted_at",
		"is_deleted",
	}
}

func (s *RoomSession) Values() []any {
	return []any{
		s.ID,
		s.SessionKey,
		s.RoomReference,
		s.RoomID,
		s.BookingID,
		s.BranchID,
		s.ExpiresAt,
		s.DeletedAt,
		s.IsDeleted,
	}
}

func (s *RoomSession) Addr() []any {
	return []any{
		&s.ID,
		&s.SessionKey,
		&s.RoomReference,
		&s.RoomID,
		&s.BookingID,
		&s.BranchID,
		&s.ExpiresAt,
		&s.DeletedAt,
		&s.IsDeleted,
		&s.CreatedAt,
		&s.UpdatedAt,
	}
}

func (s *RoomSession) IsExpired(now time.Time) bool {
	return !s.ExpiresAt.After(now)
}

func (s *RoomSession) ToResponse(roomNumber string) *RoomSessionResponse {
	return &RoomSessionResponse{
		SessionKey:    s.SessionKey,
		RoomReference: s.RoomReference,
		RoomID:        s.RoomID,
		RoomNumber:    roomNumber,
		BookingID:     s.BookingID,
		BranchID:      s.BranchID,
		ExpiresAt:     s.ExpiresAt,
	}
}
