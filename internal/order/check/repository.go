package check

import (
	"context"
	"database/sql"
	"errors"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type CheckRepository interface {
	Create(ctx context.Context, c *TableCheck) error
	GetByID(ctx context.Context, id string) (*TableCheck, error)
	GetOpenByTable(ctx context.Context, branchID string, tableNumber int) (*TableCheck, error)
	ListOpenByBranch(ctx context.Context, branchID string) ([]*TableCheck, error)
	Update(ctx context.Context, c TableCheck) error
	// ComputeBill returns subtotal (sum of non-cancelled order totals on the check) and the
	// branch merchant's tax rates so the service can derive VAT/service/total.
	ComputeBill(ctx context.Context, checkID, branchID string) (subtotal, vatPercent, serviceChargePercent float64, err error)
	// GetReadiness returns item-state counts across non-cancelled orders on the check.
	GetReadiness(ctx context.Context, checkID string) (ReadinessCounts, error)
	// GetBillPrintPolicy returns the branch merchant's bill print policy ('strict'/'lenient').
	GetBillPrintPolicy(ctx context.Context, branchID string) (string, error)
	// MarkOrdersServed moves all non-cancelled orders on the check to 'served' (called at settle).
	MarkOrdersServed(ctx context.Context, checkID string) error
}

type checkRepository struct {
	dal     *common.DAL[*TableCheck]
	joinDAL *common.JoinDAL
	logger  config.Logger
}

func NewCheckRepository(dal *common.DAL[*TableCheck], joinDAL *common.JoinDAL, logger config.Logger) CheckRepository {
	return &checkRepository{dal: dal, joinDAL: joinDAL, logger: logger}
}

func (r *checkRepository) Create(ctx context.Context, c *TableCheck) error {
	if _, err := r.dal.Create(ctx, c); err != nil {
		r.logger.Error("failed to create table check", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *checkRepository) GetByID(ctx context.Context, id string) (*TableCheck, error) {
	filter := map[string]any{"id": id, "is_deleted": false}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrCheckNotFound
		}
		r.logger.Error("failed to get table check", "error", err)
		return nil, common.ErrInternalServerError
	}
	return result, nil
}

func (r *checkRepository) GetOpenByTable(ctx context.Context, branchID string, tableNumber int) (*TableCheck, error) {
	filter := map[string]any{
		"branch_id":    branchID,
		"table_number": tableNumber,
		"status":       string(CheckOpen),
		"is_deleted":   false,
	}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrCheckNotFound
		}
		r.logger.Error("failed to get open table check", "error", err)
		return nil, common.ErrInternalServerError
	}
	return result, nil
}

func (r *checkRepository) ListOpenByBranch(ctx context.Context, branchID string) ([]*TableCheck, error) {
	filter := map[string]any{
		"branch_id":  branchID,
		"status":     string(CheckOpen),
		"is_deleted": false,
	}
	results, err := r.dal.List(ctx, filter, 1, 1000)
	if err != nil {
		r.logger.Error("failed to list open table checks", "error", err)
		return nil, common.ErrInternalServerError
	}
	return results, nil
}

func (r *checkRepository) Update(ctx context.Context, c TableCheck) error {
	filter := map[string]any{"id": c.ID, "is_deleted": false}
	updates := map[string]any{
		"status":                c.Status,
		"subtotal":              c.Subtotal,
		"vat_amount":            c.VatAmount,
		"service_charge_amount": c.ServiceChargeAmount,
		"total":                 c.Total,
		"payment_method":        c.PaymentMethod,
		"settled_by":            c.SettledBy,
		"settled_at":            c.SettledAt,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common.ErrCheckNotFound
		}
		r.logger.Error("failed to update table check", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *checkRepository) ComputeBill(ctx context.Context, checkID, branchID string) (float64, float64, float64, error) {
	const subtotalQuery = `
		SELECT COALESCE(SUM(o.total), 0)
		FROM orders o
		WHERE o.check_id = $1 AND o.is_deleted = FALSE AND o.order_status <> 'cancelled'`
	var subtotal float64
	if err := r.joinDAL.QueryRow(ctx, subtotalQuery, []any{checkID}, func(row *sql.Row) error {
		return row.Scan(&subtotal)
	}); err != nil {
		r.logger.Error("failed to compute check subtotal", "error", err)
		return 0, 0, 0, common.ErrInternalServerError
	}

	const taxQuery = `
		SELECT m.vat_percent, m.service_charge_percent
		FROM branches b
		INNER JOIN merchants m ON m.id = b.merchant_id AND m.is_deleted = FALSE
		WHERE b.id = $1 AND b.is_deleted = FALSE
		LIMIT 1`
	var vatPercent float64
	var serviceCharge sql.NullFloat64
	if err := r.joinDAL.QueryRow(ctx, taxQuery, []any{branchID}, func(row *sql.Row) error {
		return row.Scan(&vatPercent, &serviceCharge)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, 0, common.ErrBranchNotFound
		}
		r.logger.Error("failed to resolve branch tax rates", "error", err)
		return 0, 0, 0, common.ErrInternalServerError
	}
	serviceChargePercent := 0.0
	if serviceCharge.Valid {
		serviceChargePercent = serviceCharge.Float64
	}
	return subtotal, vatPercent, serviceChargePercent, nil
}

func (r *checkRepository) GetReadiness(ctx context.Context, checkID string) (ReadinessCounts, error) {
	const query = `
		SELECT
			COUNT(*) FILTER (WHERE oi.item_status <> 'cancelled') AS total_items,
			COUNT(*) FILTER (WHERE oi.item_status = 'sent') AS sent,
			COUNT(*) FILTER (WHERE oi.item_status IN ('accepted', 'preparing')) AS in_progress,
			COUNT(*) FILTER (WHERE oi.item_status = 'ready') AS ready
		FROM order_items oi
		INNER JOIN orders o ON o.id = oi.order_id AND o.is_deleted = FALSE AND o.order_status <> 'cancelled'
		WHERE o.check_id = $1 AND oi.is_deleted = FALSE`
	var counts ReadinessCounts
	if err := r.joinDAL.QueryRow(ctx, query, []any{checkID}, func(row *sql.Row) error {
		return row.Scan(&counts.TotalItems, &counts.Sent, &counts.InProgress, &counts.Ready)
	}); err != nil {
		r.logger.Error("failed to compute check readiness", "error", err)
		return ReadinessCounts{}, common.ErrInternalServerError
	}
	return counts, nil
}

func (r *checkRepository) GetBillPrintPolicy(ctx context.Context, branchID string) (string, error) {
	const query = `
		SELECT COALESCE(m.bill_print_policy, 'strict')
		FROM branches b
		INNER JOIN merchants m ON m.id = b.merchant_id AND m.is_deleted = FALSE
		WHERE b.id = $1 AND b.is_deleted = FALSE
		LIMIT 1`
	var policy string
	if err := r.joinDAL.QueryRow(ctx, query, []any{branchID}, func(row *sql.Row) error {
		return row.Scan(&policy)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return BillPrintPolicyStrict, nil
		}
		r.logger.Error("failed to resolve bill print policy", "error", err)
		return "", common.ErrInternalServerError
	}
	if policy == "" {
		policy = BillPrintPolicyStrict
	}
	return policy, nil
}

func (r *checkRepository) MarkOrdersServed(ctx context.Context, checkID string) error {
	const query = `
		UPDATE orders
		SET order_status = 'served', updated_at = NOW()
		WHERE check_id = $1 AND is_deleted = FALSE AND order_status <> 'cancelled'`
	if _, err := r.joinDAL.Exec(ctx, query, checkID); err != nil {
		r.logger.Error("failed to mark check orders served", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}
