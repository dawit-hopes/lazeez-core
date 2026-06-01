package roomsession

import "time"

type CreateRoomSessionInput struct {
	Reference string `json:"reference"`
	Passcode  string `json:"passcode"`
}

type RoomSessionResponse struct {
	SessionKey    string    `json:"session_key"`
	RoomReference string    `json:"room_reference"`
	RoomID        string    `json:"room_id"`
	RoomNumber    string    `json:"room_number"`
	BookingID     string    `json:"booking_id"`
	BranchID      string    `json:"branch_id"`
	ExpiresAt     time.Time `json:"expires_at"`
}
