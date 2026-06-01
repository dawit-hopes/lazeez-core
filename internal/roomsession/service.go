package roomsession

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"lazeez-core/internal/rooms/booking"
	"lazeez-core/internal/rooms/room"
	"time"
)

const sessionDuration = 2 * time.Hour

type RoomSessionService interface {
	Create(ctx context.Context, input CreateRoomSessionInput) (*RoomSessionResponse, error)
	GetValidForRoomOrder(ctx context.Context, sessionKey string) (*RoomSession, error)
}

type roomSessionService struct {
	repo           RoomSessionRepository
	roomRepo       room.RoomRepository
	bookingService booking.BookingService
	logger         config.Logger
}

func NewRoomSessionService(
	repo RoomSessionRepository,
	roomRepo room.RoomRepository,
	bookingService booking.BookingService,
	logger config.Logger,
) RoomSessionService {
	return &roomSessionService{
		repo:           repo,
		roomRepo:       roomRepo,
		bookingService: bookingService,
		logger:         logger,
	}
}

// Create verifies the room QR reference + passcode against the active booking and,
// on success, issues a short-lived session bound to that booking. It rate-limits
// passcode attempts to protect against brute force.
func (s *roomSessionService) Create(ctx context.Context, input CreateRoomSessionInput) (*RoomSessionResponse, error) {
	rm, err := s.roomRepo.GetByReference(ctx, input.Reference)
	if err != nil {
		return nil, err
	}

	b, err := s.bookingService.GetBookingByRoom(ctx, rm.ID)
	if err != nil {
		return nil, err
	}

	if err := s.bookingService.VerifyGuestPasscode(ctx, b, input.Passcode); err != nil {
		return nil, err
	}

	now := time.Now()
	session := &RoomSession{
		SessionKey:    common.GenerateUUID(),
		RoomReference: input.Reference,
		RoomID:        rm.ID,
		BookingID:     b.ID,
		BranchID:      rm.BranchID,
		ExpiresAt:     now.Add(sessionDuration),
	}
	session.ID = common.GenerateUUID()

	if err := s.repo.Create(ctx, session); err != nil {
		return nil, err
	}

	return session.ToResponse(rm.RoomNumber), nil
}

func (s *roomSessionService) GetValidForRoomOrder(ctx context.Context, sessionKey string) (*RoomSession, error) {
	session, err := s.repo.GetBySessionKey(ctx, sessionKey)
	if err != nil {
		return nil, err
	}
	if session.IsExpired(time.Now()) {
		return nil, common.ErrRoomSessionExpired
	}
	return session, nil
}
