package table

import (
	"database/sql"
	"lazeez-core/internal/common"
)

type Table struct {
	common.Base
	TableName     string         `db:"table_name"`
	BranchID      string         `db:"branch_id"`
	Reference     string         `db:"reference"`
	QRCode        string         `db:"qr_code"`
	QRVersion     int            `db:"qr_version"`
	Status        string         `db:"status"`
	ActiveOrderID sql.NullString `db:"active_order_id"`
}

func (t *Table) Table() string {
	return "tables"
}

func (t *Table) Columns() []string {
	return []string{"id", "table_name", "branch_id", "reference", "qr_code", "qr_version", "status", "active_order_id", "deleted_at", "is_deleted"}
}

func (t *Table) Values() []any {
	return []any{t.ID, t.TableName, t.BranchID, t.Reference, t.QRCode, t.QRVersion, t.Status, t.ActiveOrderID, t.DeletedAt, t.IsDeleted}
}

func (t *Table) Addr() []any {
	return []any{&t.ID, &t.TableName, &t.BranchID, &t.Reference, &t.QRCode, &t.QRVersion, &t.Status, &t.ActiveOrderID, &t.DeletedAt, &t.IsDeleted, &t.CreatedAt, &t.UpdatedAt}
}

func (t *Table) ToDTO() TableDTO {
	return TableDTO{
		BaseDTO: common.BaseDTO{
			ID:        t.ID,
			IsDeleted: t.IsDeleted,
			CreatedAt: t.CreatedAt,
			UpdatedAt: t.UpdatedAt,
			DeletedAt: common.ToNullTimePtr(t.DeletedAt),
		},
		TableName:         t.TableName,
		BranchID:      t.BranchID,
		Reference:     t.Reference,
		QRCode:        t.QRCode,
		QRVersion:     t.QRVersion,
		Status:        t.Status,
		ActiveOrderID: t.ActiveOrderID.String,
	}
}
