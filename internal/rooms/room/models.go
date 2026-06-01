package room

import "lazeez-core/internal/common"

type Room struct {
	common.Base
	RoomNumber string `json:"room_number" db:"room_number"`
	RoomTypeID string `json:"room_type_id" db:"room_type_id"`
	BranchID   string `json:"branch_id" db:"branch_id"`
	Floor      int    `json:"floor" db:"floor"`
	QRCode     string `json:"qr_code" db:"qr_code"`
	Status     string `json:"status" db:"status"`
	QRVersion  int    `json:"qr_version" db:"qr_version"`
	Reference  string `json:"reference" db:"reference"`
}

func (r *Room) Table() string {
	return "single_rooms"
}

func (r *Room) Columns() []string {
	return []string{"id", "room_number", "room_type_id", "branch_id", "floor", "qr_code", "qr_version", "reference", "status", "deleted_at", "is_deleted"}
}

func (r *Room) Values() []any {
		return []any{r.ID, r.RoomNumber, r.RoomTypeID, r.BranchID, r.Floor, r.QRCode, r.QRVersion, r.Reference, r.Status, r.DeletedAt, r.IsDeleted}
	}

func (r *Room) Addr() []any {
	return []any{&r.ID, &r.RoomNumber, &r.RoomTypeID, &r.BranchID, &r.Floor, &r.QRCode, &r.QRVersion, &r.Reference, &r.Status, &r.DeletedAt, &r.IsDeleted, &r.CreatedAt, &r.UpdatedAt}
}

func (r *Room) ToDTO() RoomDTO {
	return RoomDTO{
		BaseDTO:    r.Base.ToDTO(),
		RoomNumber: r.RoomNumber,
		RoomTypeID: r.RoomTypeID,
		BranchID:   r.BranchID,
		Floor:      r.Floor,
		QRCode:     r.QRCode,
		QRVersion:  r.QRVersion,
		Reference:  r.Reference,
		Status:     RoomStatus(r.Status),
	}
}
