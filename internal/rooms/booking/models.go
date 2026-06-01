package booking

import (
	"database/sql"
	"lazeez-core/internal/common"
	"time"
)

type Booking struct {
	common.Base
	RoomID              string       `json:"room_id" db:"room_id"`
	BranchID            string       `json:"branch_id" db:"branch_id"`
	GuestName           string       `json:"guest_name" db:"guest_name"`
	GuestPhone          string       `json:"guest_phone" db:"guest_phone"`
	NumberOfNights      int          `json:"number_of_nights" db:"number_of_nights"`
	CheckInDate         time.Time    `json:"check_in_date" db:"check_in_date"`
	CheckOutDate        time.Time    `json:"check_out_date" db:"check_out_date"`
	PasscodeHash        string       `json:"-" db:"passcode_hash"`
	PasscodeAttempts    int          `json:"-" db:"passcode_attempts"`
	PasscodeLockedUntil sql.NullTime `json:"-" db:"passcode_locked_until"`
	Status              string       `json:"status" db:"status"`

	// Passcode holds the plaintext guest passcode. It is only populated in the
	// check-in response and is never persisted or returned on reads.
	Passcode string `json:"-" db:"-"`
}

func (b *Booking) Table() string {
	return "bookings"
}

func (b *Booking) Columns() []string {
	return []string{"id", "room_id", "branch_id", "guest_name", "guest_phone", "number_of_nights", "check_in_date", "check_out_date", "passcode_hash", "passcode_attempts", "passcode_locked_until", "status", "deleted_at", "is_deleted"}
}

func (b *Booking) Values() []any {
	return []any{b.ID, b.RoomID, b.BranchID, b.GuestName, b.GuestPhone, b.NumberOfNights, b.CheckInDate, b.CheckOutDate, b.PasscodeHash, b.PasscodeAttempts, b.PasscodeLockedUntil, b.Status, b.DeletedAt, b.IsDeleted}
}

func (b *Booking) Addr() []any {
	return []any{&b.ID, &b.RoomID, &b.BranchID, &b.GuestName, &b.GuestPhone, &b.NumberOfNights, &b.CheckInDate, &b.CheckOutDate, &b.PasscodeHash, &b.PasscodeAttempts, &b.PasscodeLockedUntil, &b.Status, &b.DeletedAt, &b.IsDeleted, &b.CreatedAt, &b.UpdatedAt}
}

func (b *Booking) ToDTO() BookingDTO {
	return BookingDTO{
		BaseDTO: common.BaseDTO{
			ID:        b.ID,
			IsDeleted: b.IsDeleted,
			CreatedAt: b.CreatedAt,
			UpdatedAt: b.UpdatedAt,
			DeletedAt: common.ToNullTimePtr(b.DeletedAt),
		},
		RoomID:         b.RoomID,
		BranchID:       b.BranchID,
		GuestName:      b.GuestName,
		GuestPhone:     b.GuestPhone,
		NumberOfNights: b.NumberOfNights,
		CheckInDate:    b.CheckInDate,
		CheckOutDate:   b.CheckOutDate,
		Status:         BookingStatus(b.Status),
	}
}
