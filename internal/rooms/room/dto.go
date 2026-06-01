package room

import "lazeez-core/internal/common"

type RoomStatus string

const (
	RoomStatusVacant       RoomStatus = "vacant"
	RoomStatusOccupied     RoomStatus = "occupied"
	RoomStatusOutOfService RoomStatus = "out of service"
)

type RoomRequestDTO struct {
	RoomNumber string `json:"room_number"`
	RoomTypeID string `json:"room_type_id"`
	BranchID   string `json:"branch_id"`
	Floor      int    `json:"floor"`
	Status     RoomStatus `json:"status"`
}

func (r *RoomRequestDTO) ToModel() Room {
	return Room{
		RoomNumber: r.RoomNumber,
		RoomTypeID: r.RoomTypeID,
		BranchID:   r.BranchID,
		Floor:      r.Floor,
		Status:     string(r.Status),
	}
}

type RoomUpdateRequestDTO struct {
	RoomNumber string  `json:"room_number"`
	RoomTypeID string  `json:"room_type_id"`
	Floor      *int    `json:"floor"`
	Status     *RoomStatus `json:"status"`
}

type RoomDTO struct {
	common.BaseDTO
	RoomNumber string `json:"room_number"`
	RoomTypeID string `json:"room_type_id"`
	BranchID   string `json:"branch_id"`
	Floor      int    `json:"floor"`
	QRCode     string `json:"qr_code"`
	QRVersion  int    `json:"qr_version"`
	Reference  string `json:"reference"`
	Status     RoomStatus `json:"status"`
}
