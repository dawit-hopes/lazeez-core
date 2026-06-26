package check

import (
	"lazeez-core/internal/common"
	"time"
)

type CheckStatus string

const (
	CheckOpen   CheckStatus = "open"
	CheckClosed CheckStatus = "closed"
	CheckVoid   CheckStatus = "void"
)

// ReadinessState is the derived billing readiness of a check, computed from its items.
type ReadinessState string

const (
	// ReadinessNotReady means at least one item is still 'sent' (not yet accepted by a station).
	ReadinessNotReady ReadinessState = "not_ready"
	// ReadinessInProgress means every item is at least 'accepted' but not all are 'ready'.
	ReadinessInProgress ReadinessState = "in_progress"
	// ReadinessReadyForBilling means the check satisfies the merchant's bill_print_policy.
	ReadinessReadyForBilling ReadinessState = "ready_for_billing"
)

// BillPrintPolicy decides how strict the readiness gate is for settling a check.
const (
	BillPrintPolicyStrict  = "strict"
	BillPrintPolicyLenient = "lenient"
)

// ReadinessCounts is the per-state item tally used to derive readiness.
type ReadinessCounts struct {
	TotalItems int `json:"total_items"`
	Sent       int `json:"sent"`
	InProgress int `json:"in_progress"`
	Ready      int `json:"ready"`
}

type CheckDTO struct {
	common.BaseDTO
	BranchID            string      `json:"branch_id"`
	TableNumber         int         `json:"table_number"`
	Status              CheckStatus `json:"status"`
	Subtotal            float64     `json:"subtotal"`
	VatAmount           float64     `json:"vat_amount"`
	ServiceChargeAmount float64     `json:"service_charge_amount"`
	Total               float64     `json:"total"`
	PaymentMethod       string      `json:"payment_method,omitempty"`
	SettledBy           string      `json:"settled_by,omitempty"`
	SettledAt           *time.Time  `json:"settled_at,omitempty"`
}

// CheckReadinessDTO is the cashier-facing view: the bill plus its derived readiness.
type CheckReadinessDTO struct {
	CheckDTO
	Readiness ReadinessState  `json:"readiness"`
	Counts    ReadinessCounts `json:"counts"`
}

// SettleCheckInput is the cashier request to settle a check (cash-only for now).
type SettleCheckInput struct {
	PaymentMethod string `json:"payment_method"`
}
