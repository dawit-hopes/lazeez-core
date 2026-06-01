package roomsession

import (
	"context"
	"database/sql"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"lazeez-core/internal/key"
	"lazeez-core/internal/rooms/booking"
	"lazeez-core/internal/rooms/room"
	"time"
)

const (
	sessionDuration     = 2 * time.Hour
	maxPasscodeAttempts = 5
	lockoutDuration     = 15 * time.Minute
)

type RoomSessionService interface {
	Create(ctx context.Context, input CreateRoomSessionInput) (*RoomSessionResponse, error)
	GetValidForRoomOrder(ctx context.Context, sessionKey string) (*RoomSession, error)
}

type roomSessionService struct {
	repo        RoomSessionRepository
	roomRepo    room.RoomRepository
	bookingRepo booking.BookingRepository
	keyService  key.KeyService
	logger      config.Logger
}

func NewRoomSessionService(
	repo RoomSessionRepository,
	roomRepo room.RoomRepository,
	bookingRepo booking.BookingRepository,
	keyService key.KeyService,
	logger config.Logger,
) RoomSessionService {
	return &roomSessionService{
		repo:        repo,
		roomRepo:    roomRepo,
		bookingRepo: bookingRepo,
		keyService:  keyService,
		logger:      logger,
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

	b, err := s.bookingRepo.GetActiveByRoomID(ctx, rm.ID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if b.PasscodeLockedUntil.Valid && b.PasscodeLockedUntil.Time.After(now) {
		s.logger.Warn("passcode verification locked", "booking_id", b.ID, "room_id", rm.ID)
		return nil, common.ErrPasscodeLocked
	}

	ok, _ := s.keyService.VerifyPassword(input.Passcode, b.PasscodeHash)
	if !ok {
		return nil, s.handleFailedAttempt(ctx, b, now)
	}

	// Successful verification: clear any prior failed attempts / lockout.
	if b.PasscodeAttempts != 0 || b.PasscodeLockedUntil.Valid {
		if resetErr := s.bookingRepo.SetPasscodeAttempts(ctx, b.ID, 0, sql.NullTime{}); resetErr != nil {
			s.logger.Error("failed to reset passcode attempts", "booking_id", b.ID, "error", resetErr)
		}
	}

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

// handleFailedAttempt increments the failed-attempt counter and locks the booking
// once the threshold is reached. It returns the error to surface to the caller.
func (s *roomSessionService) handleFailedAttempt(ctx context.Context, b *booking.Booking, now time.Time) error {
	attempts := b.PasscodeAttempts + 1
	var locked sql.NullTime
	if attempts >= maxPasscodeAttempts {
		locked = sql.NullTime{Time: now.Add(lockoutDuration), Valid: true}
		attempts = 0 // reset counter; the lockout window now gates further attempts
	}
	if setErr := s.bookingRepo.SetPasscodeAttempts(ctx, b.ID, attempts, locked); setErr != nil {
		s.logger.Error("failed to record passcode attempt", "booking_id", b.ID, "error", setErr)
	}
	if locked.Valid {
		return common.ErrPasscodeLocked
	}
	return common.ErrInvalidPasscode
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
