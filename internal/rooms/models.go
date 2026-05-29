package rooms

import (
	"database/sql"
	"lazeez-core/internal/common"
)

type Room struct {
	common.Base
	Name          string         `json:"name" db:"name"`
	Description   string         `json:"description" db:"description"`
	MerchantID    string         `json:"merchant_id" db:"merchant_id"`
	BranchID      sql.NullString `json:"branch_id" db:"branch_id"`
	ParentID      sql.NullString `json:"parent_id" db:"parent_id"`
	PricePerNight float64        `json:"price_per_night" db:"price_per_night"`
}

func (r *Room) Table() string {
	return "rooms"
}

func (r *Room) Columns() []string {
	return []string{
		"id", "name", "description", "merchant_id", "branch_id", "parent_id",
		"price_per_night", "deleted_at", "is_deleted",
	}
}

func (r *Room) Values() []any {
	return []any{
		r.ID, r.Name, r.Description, r.MerchantID, r.BranchID, r.ParentID,
		r.PricePerNight, r.DeletedAt, r.IsDeleted,
	}
}

func (r *Room) Addr() []any {
	return []any{
		&r.ID, &r.Name, &r.Description, &r.MerchantID, &r.BranchID, &r.ParentID,
		&r.PricePerNight, &r.DeletedAt, &r.IsDeleted, &r.CreatedAt, &r.UpdatedAt,
	}
}

func (r *Room) BranchIDString() string {
	if r.BranchID.Valid {
		return r.BranchID.String
	}
	return ""
}

func (r *Room) ParentIDString() string {
	if r.ParentID.Valid {
		return r.ParentID.String
	}
	return ""
}

func (r *Room) IsMaster() bool {
	return !r.BranchID.Valid && !r.ParentID.Valid
}

func (r *Room) IsClone() bool {
	return r.ParentID.Valid && r.ParentID.String != ""
}

func (r *Room) ToDTO() RoomDTO {
	dto := RoomDTO{
		BaseDTO:       r.Base.ToDTO(),
		Name:          r.Name,
		Description:   r.Description,
		MerchantID:    r.MerchantID,
		BranchID:      r.BranchIDString(),
		ParentID:      r.ParentIDString(),
		PricePerNight: r.PricePerNight,
		IsMaster: r.IsMaster(),
		IsClone:  r.IsClone(),
	}
	return dto
}
