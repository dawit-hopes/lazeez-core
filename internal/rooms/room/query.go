package room

import (
	"database/sql"
)

const roomWithTypeSelect = `
SELECT
	sr.id, sr.room_number, sr.room_type_id, sr.branch_id, sr.floor,
	sr.qr_code, sr.qr_version, sr.reference, sr.status,
	sr.deleted_at, sr.is_deleted, sr.created_at, sr.updated_at,
	rt.id, rt.name, rt.description, rt.merchant_id, rt.branch_id, rt.parent_id,
	rt.price_per_night,
	(rt.branch_id IS NULL AND rt.parent_id IS NULL) AS type_is_master,
	(rt.parent_id IS NOT NULL) AS type_is_clone
FROM single_rooms sr
INNER JOIN rooms rt ON rt.id = sr.room_type_id AND rt.is_deleted = FALSE`

func scanRoomWithType(row interface {
	Scan(dest ...any) error
}) (*RoomWithType, error) {
	var rm Room
	var typeID, typeName, typeDescription, typeMerchantID string
	var typeBranchID, typeParentID sql.NullString
	var typePrice float64
	var typeIsMaster, typeIsClone bool

	if err := row.Scan(
		&rm.ID, &rm.RoomNumber, &rm.RoomTypeID, &rm.BranchID, &rm.Floor,
		&rm.QRCode, &rm.QRVersion, &rm.Reference, &rm.Status,
		&rm.DeletedAt, &rm.IsDeleted, &rm.CreatedAt, &rm.UpdatedAt,
		&typeID, &typeName, &typeDescription, &typeMerchantID, &typeBranchID, &typeParentID,
		&typePrice, &typeIsMaster, &typeIsClone,
	); err != nil {
		return nil, err
	}

	branchID := ""
	if typeBranchID.Valid {
		branchID = typeBranchID.String
	}
	parentID := ""
	if typeParentID.Valid {
		parentID = typeParentID.String
	}

	return &RoomWithType{
		Room: rm,
		RoomType: RoomTypeInfo{
			ID:            typeID,
			Name:          typeName,
			Description:   typeDescription,
			MerchantID:    typeMerchantID,
			BranchID:      branchID,
			ParentID:      parentID,
			PricePerNight: typePrice,
			IsMaster:      typeIsMaster,
			IsClone:       typeIsClone,
		},
	}, nil
}
