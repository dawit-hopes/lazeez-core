package booking

import (
	"context"
	"database/sql"
	"lazeez-core/internal/common"
	"time"
)

const MaxPasscodeAttempts = 5

// VerifyGuestPasscode checks lockout, compares the passcode to the booking hash,
// increments failed attempts (locking after MaxPasscodeAttempts), and resets on success.
// The same booking-level counter is used for room sessions and room orders.
// Lock persists until front desk regenerates the passcode (passcode_locked_until set).
func (s *bookingService) VerifyGuestPasscode(ctx context.Context, b *Booking, passcode string) error {
	if b.PasscodeLockedUntil.Valid {
		s.logger.Warn("passcode verification locked", "booking_id", b.ID, "room_id", b.RoomID)
		return common.ErrPasscodeLocked
	}

	ok, _ := s.keyService.VerifyPassword(passcode, b.PasscodeHash)
	if !ok {
		return s.recordFailedPasscodeAttempt(ctx, b, time.Now())
	}

	if b.PasscodeAttempts != 0 || b.PasscodeLockedUntil.Valid {
		if resetErr := s.repository.SetPasscodeAttempts(ctx, b.ID, 0, sql.NullTime{}); resetErr != nil {
			s.logger.Error("failed to reset passcode attempts", "booking_id", b.ID, "error", resetErr)
		}
	}
	return nil
}

func (s *bookingService) recordFailedPasscodeAttempt(ctx context.Context, b *Booking, now time.Time) error {
	attempts := b.PasscodeAttempts + 1
	var locked sql.NullTime
	if attempts >= MaxPasscodeAttempts {
		locked = sql.NullTime{Time: now, Valid: true}
		attempts = 0
	}
	if setErr := s.repository.SetPasscodeAttempts(ctx, b.ID, attempts, locked); setErr != nil {
		s.logger.Error("failed to record passcode attempt", "booking_id", b.ID, "error", setErr)
	}
	if locked.Valid {
		return common.ErrPasscodeLocked
	}
	return common.ErrInvalidPasscode
}
