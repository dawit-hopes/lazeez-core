package table

import "lazeez-core/internal/common"

type TableDTO struct {
	common.BaseDTO
	TableName     string `json:"table_name"`
	BranchID      string `json:"branch_id"`
	Reference     string `json:"reference"`
	QRCode        string `json:"qr_code"`
	QRVersion     int    `json:"qr_version"`
	Status        string `json:"status"`
	ActiveOrderID string `json:"active_order_id"`
}

type TableQR struct {
	TableName string `json:"table"`
}

type TableRequest struct {
	BranchID string    `json:"branch_id"`
	Tables   []TableQR `json:"tables"`
}

// AttachOrderRequest is the body for POST /tables/{id}/attach-order
type AttachOrderRequest struct {
	OrderID string `json:"order_id"`
}
