package rooms

import "lazeez-core/internal/common"

type RoomRequestDTO struct {
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	MerchantID    string  `json:"merchant_id"`
	BranchID      string  `json:"branch_id"`
	PricePerNight float64 `json:"price_per_night"`
}

type CloneRoomRequestDTO struct {
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	PricePerNight float64 `json:"price_per_night"`
}

type RoomDTO struct {
	common.BaseDTO
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	MerchantID    string  `json:"merchant_id"`
	BranchID      string  `json:"branch_id,omitempty"`
	ParentID      string  `json:"parent_id,omitempty"`
	PricePerNight float64 `json:"price_per_night"`
	IsMaster      bool    `json:"is_master"`
	IsClone       bool    `json:"is_clone"`
	CanEdit       bool    `json:"can_edit"`
}

func (r *RoomRequestDTO) ToModel() Room {
	return Room{
		Name:          r.Name,
		Description:   r.Description,
		MerchantID:    r.MerchantID,
		BranchID:      common.ToNUllString(r.BranchID),
		PricePerNight: r.PricePerNight,
	}
}
