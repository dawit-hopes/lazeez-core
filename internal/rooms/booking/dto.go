package booking

import (
	"lazeez-core/internal/common"
	"time"
)

type BookingStatus string

const (
	BookingActive     BookingStatus = "active"
	BookingCheckedOut BookingStatus = "checked_out"
	BookingCancelled  BookingStatus = "cancelled"
)

// BookingRequestDTO is the check-in payload. branch_id, passcode and status are
// set server-side: branch is derived from the room, passcode is generated, and
// status is always "active" on check-in.
type BookingRequestDTO struct {
	RoomID         string    `json:"room_id"`
	GuestName      string    `json:"guest_name"`
	GuestPhone     string    `json:"guest_phone"`
	NumberOfNights int       `json:"number_of_nights"`
	CheckInDate    common.Date `json:"check_in_date"`
}

func (b *BookingRequestDTO) ToModel() Booking {
	return Booking{
		RoomID:         b.RoomID,
		GuestName:      b.GuestName,
		GuestPhone:     b.GuestPhone,
		NumberOfNights: b.NumberOfNights,
		CheckInDate:    b.CheckInDate.Time(),
	}
}

// BookingUpdateRequestDTO allows editing guest details on an existing booking.
type BookingUpdateRequestDTO struct {
	GuestName  string `json:"guest_name"`
	GuestPhone string `json:"guest_phone"`
}

func (b *BookingUpdateRequestDTO) IsEmpty() bool {
	return b.GuestName == "" && b.GuestPhone == ""
}

type BookingDTO struct {
	common.BaseDTO
	RoomID         string        `json:"room_id"`
	BranchID       string        `json:"branch_id"`
	GuestName      string        `json:"guest_name"`
	GuestPhone     string        `json:"guest_phone"`
	NumberOfNights int           `json:"number_of_nights"`
	CheckInDate    time.Time     `json:"check_in_date"`
	CheckOutDate   time.Time     `json:"check_out_date"`
	// Passcode is the plaintext guest passcode, returned only once in the
	// check-in response. Reads (Get/List) never include it.
	Passcode string        `json:"passcode,omitempty"`
	Status   BookingStatus `json:"status"`
}
