package order

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"lazeez-core/config"
	"lazeez-core/internal/common"
	item "lazeez-core/internal/order/Item"
)

type OrderRepository interface {
	Create(ctx context.Context, order Order) (Order, error)
	Get(ctx context.Context, id string) (*OrderDTO, error)
	GetBySessionKey(ctx context.Context, id string, sessionKey string) (*OrderDTO, error)
	Update(ctx context.Context, order Order) error
	UpdateStatus(ctx context.Context, id string, status string, cancellationReason, paymentTransactionID string) error
	UpdatePaymentStatus(ctx context.Context, id string, input PaymentStatusUpdate) error
	Delete(ctx context.Context, id string) error
	UnDelete(ctx context.Context, id string) error
	ListBySessionKey(ctx context.Context, filter OrderFilter, sessionKey string) (*common.PaginatedResponse[[]*OrderDTO], error)
	ListByBranch(ctx context.Context, filter OrderFilter, branchID string) (*common.PaginatedResponse[[]*OrderDTO], error)
	ListArchiveByBranch(ctx context.Context, filter OrderFilter, branchID string) (*common.PaginatedResponse[[]*OrderDTO], error)
	ListAll(ctx context.Context, filter OrderFilter) (*common.PaginatedResponse[[]*OrderDTO], error)
	CheckExists(ctx context.Context, tableNumber int, branchID string) error
	GetBranchIDByReference(ctx context.Context, reference string) (string, error)
	OrderNumberExists(ctx context.Context, branchID string, orderNumber int) (bool, error)
}

type orderRepository struct {
	dal     *common.DAL[*Order]
	joinDAL *common.JoinDAL
	logger  config.Logger
}

func NewOrderRepository(dal *common.DAL[*Order], joinDAL *common.JoinDAL, logger config.Logger) OrderRepository {
	return &orderRepository{dal: dal, joinDAL: joinDAL, logger: logger}
}

// orderWithRelations selects orders with order items via LATERAL join for faster execution.
const orderWithRelations = `
SELECT 
	o.id, o.order_number, o.table_number, o.branch_id, o.session_key, o.order_status, o.cancellation_reason, o.total, o.payment_method, o.payment_status,
	o.payment_date, o.payment_amount, o.payment_currency, o.payment_transaction_id,
	o.created_at, o.updated_at,
	COALESCE(items.agg, '[]'::json)::text AS order_items
FROM orders o
LEFT JOIN LATERAL (
	SELECT json_agg(json_build_object(
		'id', oi.id, 'menu_item_id', oi.menu_item_id, 'name', m.name,
		'modifier_options', oi.modifier_options,
		'quantity', oi.quantity, 'price', oi.price, 'total', oi.total
	)) AS agg
	FROM order_items oi
	LEFT JOIN menus m ON m.id = oi.menu_item_id AND m.is_deleted = FALSE
	WHERE oi.order_id = o.id AND oi.is_deleted = FALSE
) items ON true
WHERE o.is_deleted = FALSE
`

func (r *orderRepository) Create(ctx context.Context, order Order) (Order, error) {
	created, err := r.dal.Create(ctx, &order)
	if err != nil {
		r.logger.Error("failed to create order", "error", err)
		return Order{}, common.ErrInternalServerError
	}
	return *created, nil
}

func (r *orderRepository) Get(ctx context.Context, id string) (*OrderDTO, error) {
	query := orderWithRelations + " AND o.id = $1"
	var dto OrderDTO
	err := r.joinDAL.QueryRow(ctx, query, []any{id}, func(row *sql.Row) error {
		return r.scanOrderWithRelations(row, &dto)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("order not found", "error", err)
			return nil, common.ErrOrderNotFound
		}
		r.logger.Error("failed to get order", "error", err)
		return nil, common.ErrInternalServerError
	}
	return &dto, nil
}

func (r *orderRepository) GetBySessionKey(ctx context.Context, id string, sessionKey string) (*OrderDTO, error) {
	query := orderWithRelations + " AND o.id = $1 AND o.session_key = $2"
	var dto OrderDTO
	err := r.joinDAL.QueryRow(ctx, query, []any{id, sessionKey}, func(row *sql.Row) error {
		return r.scanOrderWithRelations(row, &dto)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrOrderNotFound
		}
		r.logger.Error("failed to get order by session key", "error", err)
		return nil, common.ErrInternalServerError
	}
	return &dto, nil
}

func (r *orderRepository) Update(ctx context.Context, order Order) error {
	filter := map[string]any{"id": order.ID}
	updates := map[string]any{
		"order_number":           order.OrderNumber,
		"table_number":           order.TableNumber,
		"branch_id":              order.BranchID,
		"order_status":           order.OrderStatus,
		"total":                  order.Total,
		"payment_method":         order.PaymentMethod,
		"payment_status":         order.PaymentStatus,
		"payment_date":           order.PaymentDate,
		"payment_amount":         order.PaymentAmount,
		"payment_currency":       order.PaymentCurrency,
		"payment_transaction_id": order.PaymentTransactionID,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("order not found", "error", err)
			return common.ErrOrderNotFound
		}
		r.logger.Error("failed to update order", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

type PaymentStatusUpdate struct {
	PaymentStatus        string
	PaymentTransactionID string
	PaymentAmount        float64
	PaymentCurrency      string
	OrderStatus          string
	CancellationReason   string
}

func (r *orderRepository) UpdatePaymentStatus(ctx context.Context, id string, input PaymentStatusUpdate) error {
	filter := map[string]any{"id": id}
	updates := map[string]any{
		"payment_status": input.PaymentStatus,
		"payment_date":   time.Now(),
	}
	if input.PaymentTransactionID != "" {
		updates["payment_transaction_id"] = input.PaymentTransactionID
	}
	if input.PaymentAmount > 0 {
		updates["payment_amount"] = input.PaymentAmount
	}
	if input.PaymentCurrency != "" {
		updates["payment_currency"] = input.PaymentCurrency
	}
	if input.OrderStatus != "" {
		updates["order_status"] = input.OrderStatus
	}
	if input.CancellationReason != "" {
		updates["cancellation_reason"] = input.CancellationReason
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("order not found", "error", err)
			return common.ErrOrderNotFound
		}
		r.logger.Error("failed to update payment status", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id string, status string, cancellationReason, paymentTransactionID string) error {
	filter := map[string]any{"id": id}
	updates := map[string]any{"order_status": status}
	if cancellationReason != "" {
		updates["cancellation_reason"] = cancellationReason
	}
	if paymentTransactionID != "" {
		updates["payment_transaction_id"] = paymentTransactionID
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("order not found", "error", err)
			return common.ErrOrderNotFound
		}
		r.logger.Error("failed to update order status", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *orderRepository) Delete(ctx context.Context, id string) error {
	filter := map[string]any{"id": id}
	updates := map[string]any{
		"is_deleted": true,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		r.logger.Error("failed to delete order", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *orderRepository) UnDelete(ctx context.Context, id string) error {
	filter := map[string]any{"id": id}
	updates := map[string]any{
		"is_deleted": false,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		r.logger.Error("failed to undelete order", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *orderRepository) ListByBranch(ctx context.Context, filter OrderFilter, branchID string) (*common.PaginatedResponse[[]*OrderDTO], error) {
	return r.listOrders(ctx, filter, " AND o.branch_id = $1 ", []any{branchID}, "o.created_at DESC")
}

func (r *orderRepository) ListArchiveByBranch(ctx context.Context, filter OrderFilter, branchID string) (*common.PaginatedResponse[[]*OrderDTO], error) {
	baseCond := " AND o.branch_id = $1 AND o.order_status <> '" + string(StatusPending) + "' "
	return r.listOrdersWithCount(ctx, filter, baseCond, []any{branchID}, BuildArchiveOrderFilterClause, "o.updated_at DESC")
}

func (r *orderRepository) ListAll(ctx context.Context, filter OrderFilter) (*common.PaginatedResponse[[]*OrderDTO], error) {
	return r.listOrders(ctx, filter, " ", []any{}, "o.created_at DESC")
}

func (r *orderRepository) listOrders(ctx context.Context, filter OrderFilter, baseCond string, baseArgs []any, orderBy string) (*common.PaginatedResponse[[]*OrderDTO], error) {
	filterCond, args := BuildOrderFilterClause(filter, baseArgs)
	offset := (filter.Page - 1) * filter.Limit
	argNum := len(args) + 1
	query := orderWithRelations + baseCond + filterCond + " ORDER BY " + orderBy + " LIMIT $" + strconv.Itoa(argNum) + " OFFSET $" + strconv.Itoa(argNum+1)
	args = append(args, filter.Limit, offset)

	results, err := common.QueryRows(r.joinDAL, ctx, query, args, func(rows *sql.Rows) (*OrderDTO, error) {
		var dto OrderDTO
		if err := r.scanOrderWithRelationsFromRows(rows, &dto); err != nil {
			return nil, err
		}
		return &dto, nil
	})
	if err != nil {
		r.logger.Error("failed to list orders", "error", err)
		return nil, common.ErrInternalServerError
	}
	return &common.PaginatedResponse[[]*OrderDTO]{
		Data: results,
		Meta: common.BuildPaginationMeta(int64(len(results)), filter.Page, filter.Limit),
	}, nil
}

type orderFilterClauseBuilder func(OrderFilter, []any) (string, []any)

func (r *orderRepository) listOrdersWithCount(
	ctx context.Context,
	filter OrderFilter,
	baseCond string,
	baseArgs []any,
	buildClause orderFilterClauseBuilder,
	orderBy string,
) (*common.PaginatedResponse[[]*OrderDTO], error) {
	filterCond, args := buildClause(filter, baseArgs)

	countQuery := "SELECT COUNT(*) FROM orders o WHERE o.is_deleted = FALSE" + baseCond + filterCond
	var total int64
	err := r.joinDAL.QueryRow(ctx, countQuery, args, func(row *sql.Row) error {
		return row.Scan(&total)
	})
	if err != nil {
		r.logger.Error("failed to count orders", "error", err)
		return nil, common.ErrInternalServerError
	}

	offset := (filter.Page - 1) * filter.Limit
	argNum := len(args) + 1
	query := orderWithRelations + baseCond + filterCond + " ORDER BY " + orderBy + " LIMIT $" + strconv.Itoa(argNum) + " OFFSET $" + strconv.Itoa(argNum+1)
	listArgs := append(append([]any{}, args...), filter.Limit, offset)

	results, err := common.QueryRows(r.joinDAL, ctx, query, listArgs, func(rows *sql.Rows) (*OrderDTO, error) {
		var dto OrderDTO
		if err := r.scanOrderWithRelationsFromRows(rows, &dto); err != nil {
			return nil, err
		}
		return &dto, nil
	})
	if err != nil {
		r.logger.Error("failed to list orders", "error", err)
		return nil, common.ErrInternalServerError
	}

	return &common.PaginatedResponse[[]*OrderDTO]{
		Data: results,
		Meta: common.BuildPaginationMeta(total, filter.Page, filter.Limit),
	}, nil
}

func (r *orderRepository) ListBySessionKey(ctx context.Context, filter OrderFilter, sessionKey string) (*common.PaginatedResponse[[]*OrderDTO], error) {
	return r.listOrders(ctx, filter, " AND o.session_key = $1 ", []any{sessionKey}, "o.created_at DESC")
}

func (r *orderRepository) OrderNumberExists(ctx context.Context, branchID string, orderNumber int) (bool, error) {
	filter := map[string]any{
		"branch_id":     branchID,
		"order_number":  orderNumber,
		"is_deleted":    false,
	}
	_, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		r.logger.Error("failed to check order number", "error", err)
		return false, common.ErrInternalServerError
	}
	return true, nil
}

func (r *orderRepository) GetBranchIDByReference(ctx context.Context, reference string) (string, error) {
	const query = `SELECT branch_id::text FROM tables WHERE reference = $1 AND is_deleted = FALSE LIMIT 1`
	var branchID string
	err := r.joinDAL.QueryRow(ctx, query, []any{reference}, func(row *sql.Row) error {
		return row.Scan(&branchID)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("table reference not found", "reference", reference)
			return "", common.ErrReferenceNotValid
		}
		r.logger.Error("failed to resolve branch by reference", "error", err)
		return "", common.ErrInternalServerError
	}
	return branchID, nil
}

func (r *orderRepository) CheckExists(ctx context.Context, tableNumber int, branchID string) error {
	filter := map[string]any{"table_number": tableNumber, "branch_id": branchID, "is_deleted": false}
	_, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		r.logger.Error("failed to check if order exists", "error", err)
		return common.ErrInternalServerError
	}
	return common.ErrOrderAlreadyExists
}

func (r *orderRepository) scanOrderWithRelations(row *sql.Row, dto *OrderDTO) error {
	var orderItemsJSON []byte
	var cancellationReason sql.NullString
	err := row.Scan(
		&dto.ID, &dto.OrderNumber, &dto.TableNumber, &dto.BranchID, &dto.SessionKey, &dto.OrderStatus, &cancellationReason, &dto.Total,
		&dto.PaymentMethod, &dto.PaymentStatus, &dto.PaymentDate, &dto.PaymentAmount,
		&dto.PaymentCurrency, &dto.PaymentTransactionID,
		&dto.CreatedAt, &dto.UpdatedAt,
		&orderItemsJSON,
	)
	if err != nil {
		return err
	}
	if cancellationReason.Valid {
		dto.CancellationReason = cancellationReason.String
	}
	return r.unmarshalOrderItems(dto, orderItemsJSON)
}

func (r *orderRepository) scanOrderWithRelationsFromRows(rows *sql.Rows, dto *OrderDTO) error {
	var orderItemsJSON []byte
	var cancellationReason sql.NullString
	err := rows.Scan(
		&dto.ID, &dto.OrderNumber, &dto.TableNumber, &dto.BranchID, &dto.SessionKey, &dto.OrderStatus, &cancellationReason, &dto.Total,
		&dto.PaymentMethod, &dto.PaymentStatus, &dto.PaymentDate, &dto.PaymentAmount,
		&dto.PaymentCurrency, &dto.PaymentTransactionID,
		&dto.CreatedAt, &dto.UpdatedAt,
		&orderItemsJSON,
	)
	if err != nil {
		return err
	}
	if cancellationReason.Valid {
		dto.CancellationReason = cancellationReason.String
	}
	return r.unmarshalOrderItems(dto, orderItemsJSON)
}

func (r *orderRepository) unmarshalOrderItems(dto *OrderDTO, orderItemsJSON []byte) error {
	var items []*item.OrderItemDTO
	if err := json.Unmarshal(orderItemsJSON, &items); err != nil {
		return err
	}
	if items == nil {
		items = []*item.OrderItemDTO{}
	}
	dto.OrderItems = make([]item.OrderItemDTO, len(items))
	for i, it := range items {
		if it != nil {
			dto.OrderItems[i] = *it
		}
	}
	return nil
}
