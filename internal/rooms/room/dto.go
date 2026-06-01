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

// RoomTypeInfo is the room type (rooms table) embedded on physical room responses.
type RoomTypeInfo struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	MerchantID    string  `json:"merchant_id"`
	BranchID      string  `json:"branch_id,omitempty"`
	ParentID      string  `json:"parent_id,omitempty"`
	PricePerNight float64 `json:"price_per_night"`
	IsMaster      bool    `json:"is_master"`
	IsClone       bool    `json:"is_clone"`
}

type RoomDTO struct {
	common.BaseDTO
	RoomNumber string     `json:"room_number"`
	RoomType   RoomTypeInfo `json:"room_type"`
	BranchID   string     `json:"branch_id"`
	Floor      int        `json:"floor"`
	QRCode     string     `json:"qr_code"`
	QRVersion  int        `json:"qr_version"`
	Reference  string     `json:"reference"`
	Status     RoomStatus `json:"status"`
}

// RoomWithType is a physical room row joined with its room type.
type RoomWithType struct {
	Room
	RoomType RoomTypeInfo
}

func (r *RoomWithType) ToDTO() RoomDTO {
	return RoomDTO{
		BaseDTO:    r.Room.Base.ToDTO(),
		RoomNumber: r.Room.RoomNumber,
		RoomType:   r.RoomType,
		BranchID:   r.Room.BranchID,
		Floor:      r.Room.Floor,
		QRCode:     r.Room.QRCode,
		QRVersion:  r.Room.QRVersion,
		Reference:  r.Room.Reference,
		Status:     RoomStatus(r.Room.Status),
	}
}
