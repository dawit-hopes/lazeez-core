package folio

import (
	"lazeez-core/internal/common"
	"time"
)

type BillStatus string

const (
	BillOpen    BillStatus = "open"
	BillSettled BillStatus = "settled"
	BillVoid    BillStatus = "void"
)

type BillDTO struct {
	common.BaseDTO
	BookingID     string     `json:"booking_id"`
	BranchID      string     `json:"branch_id"`
	Status        BillStatus `json:"status"`
	Total         float64    `json:"total"`
	PaymentMethod string     `json:"payment_method"`
	SettledBy     string     `json:"settled_by,omitempty"`
	SettledAt     *time.Time `json:"settled_at,omitempty"`
}

// SettleBillInput is the front-desk/room-service request to settle a folio.
type SettleBillInput struct {
	PaymentMethod string `json:"payment_method"`
}
