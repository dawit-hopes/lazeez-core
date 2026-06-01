package roomorder

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"

	"lazeez-core/config"
	"lazeez-core/internal/common"
	item "lazeez-core/internal/order/Item"
)

type RoomOrderRepository interface {
	Create(ctx context.Context, order RoomOrder) (RoomOrder, error)
	CreateItem(ctx context.Context, orderItem RoomOrderItem) error
	Get(ctx context.Context, id string) (*RoomOrderDTO, error)
	GetBySessionKey(ctx context.Context, id, sessionKey string) (*RoomOrderDTO, error)
	UpdateStatus(ctx context.Context, id, status, cancellationReason string) error
	Delete(ctx context.Context, id string) error
	ListBySessionKey(ctx context.Context, filter RoomOrderFilter, sessionKey string) (*common.PaginatedResponse[[]*RoomOrderDTO], error)
	ListByBranch(ctx context.Context, filter RoomOrderFilter, branchID string) (*common.PaginatedResponse[[]*RoomOrderDTO], error)
	ListArchiveByBranch(ctx context.Context, filter RoomOrderFilter, branchID string) (*common.PaginatedResponse[[]*RoomOrderDTO], error)
	ListAll(ctx context.Context, filter RoomOrderFilter) (*common.PaginatedResponse[[]*RoomOrderDTO], error)
	OrderNumberExists(ctx context.Context, branchID string, orderNumber int) (bool, error)
	SumActiveByBooking(ctx context.Context, bookingID string) (float64, error)
}

type roomOrderRepository struct {
	dal     *common.DAL[*RoomOrder]
	itemDAL *common.DAL[*RoomOrderItem]
	joinDAL *common.JoinDAL
	logger  config.Logger
}

func NewRoomOrderRepository(dal *common.DAL[*RoomOrder], itemDAL *common.DAL[*RoomOrderItem], joinDAL *common.JoinDAL, logger config.Logger) RoomOrderRepository {
	return &roomOrderRepository{dal: dal, itemDAL: itemDAL, joinDAL: joinDAL, logger: logger}
}

const roomOrderWithRelations = `
SELECT
	o.id, o.order_number, o.room_id, o.booking_id, o.branch_id, o.session_key, o.order_status, o.cancellation_reason, o.total, o.bill_id,
	o.created_at, o.updated_at,
	COALESCE(items.agg, '[]'::json)::text AS order_items
FROM room_orders o
LEFT JOIN LATERAL (
	SELECT json_agg(json_build_object(
		'id', oi.id, 'menu_item_id', oi.menu_item_id, 'name', m.name,
		'modifier_options', oi.modifier_options,
		'quantity', oi.quantity, 'price', oi.price, 'total', oi.total
	)) AS agg
	FROM room_order_items oi
	LEFT JOIN menus m ON m.id = oi.menu_item_id AND m.is_deleted = FALSE
	WHERE oi.room_order_id = o.id AND oi.is_deleted = FALSE
) items ON true
WHERE o.is_deleted = FALSE
`

func (r *roomOrderRepository) Create(ctx context.Context, order RoomOrder) (RoomOrder, error) {
	created, err := r.dal.Create(ctx, &order)
	if err != nil {
		r.logger.Error("failed to create room order", "error", err)
		return RoomOrder{}, common.ErrInternalServerError
	}
	return *created, nil
}

func (r *roomOrderRepository) CreateItem(ctx context.Context, orderItem RoomOrderItem) error {
	if orderItem.ModifierOptions == nil {
		orderItem.ModifierOptions = []string{}
	}
	if _, err := r.itemDAL.Create(ctx, &orderItem); err != nil {
		r.logger.Error("failed to create room order item", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *roomOrderRepository) Get(ctx context.Context, id string) (*RoomOrderDTO, error) {
	query := roomOrderWithRelations + " AND o.id = $1"
	var dto RoomOrderDTO
	err := r.joinDAL.QueryRow(ctx, query, []any{id}, func(row *sql.Row) error {
		return r.scanOrderRow(row, &dto)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrOrderNotFound
		}
		r.logger.Error("failed to get room order", "error", err)
		return nil, common.ErrInternalServerError
	}
	return &dto, nil
}

func (r *roomOrderRepository) GetBySessionKey(ctx context.Context, id, sessionKey string) (*RoomOrderDTO, error) {
	query := roomOrderWithRelations + " AND o.id = $1 AND o.session_key = $2"
	var dto RoomOrderDTO
	err := r.joinDAL.QueryRow(ctx, query, []any{id, sessionKey}, func(row *sql.Row) error {
		return r.scanOrderRow(row, &dto)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrOrderNotFound
		}
		r.logger.Error("failed to get room order by session key", "error", err)
		return nil, common.ErrInternalServerError
	}
	return &dto, nil
}

func (r *roomOrderRepository) UpdateStatus(ctx context.Context, id, status, cancellationReason string) error {
	filter := map[string]any{"id": id}
	updates := map[string]any{"order_status": status}
	if cancellationReason != "" {
		updates["cancellation_reason"] = cancellationReason
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common.ErrOrderNotFound
		}
		r.logger.Error("failed to update room order status", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *roomOrderRepository) Delete(ctx context.Context, id string) error {
	filter := map[string]any{"id": id}
	updates := map[string]any{"is_deleted": true}
	if err := r.dal.Update(ctx, filter, updates); err != nil {
		r.logger.Error("failed to delete room order", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *roomOrderRepository) ListBySessionKey(ctx context.Context, filter RoomOrderFilter, sessionKey string) (*common.PaginatedResponse[[]*RoomOrderDTO], error) {
	return r.listOrders(ctx, filter, " AND o.session_key = $1 ", []any{sessionKey}, "o.created_at DESC")
}

func (r *roomOrderRepository) ListByBranch(ctx context.Context, filter RoomOrderFilter, branchID string) (*common.PaginatedResponse[[]*RoomOrderDTO], error) {
	return r.listOrders(ctx, filter, " AND o.branch_id = $1 ", []any{branchID}, "o.created_at DESC")
}

func (r *roomOrderRepository) ListArchiveByBranch(ctx context.Context, filter RoomOrderFilter, branchID string) (*common.PaginatedResponse[[]*RoomOrderDTO], error) {
	baseCond := " AND o.branch_id = $1 AND o.order_status <> '" + string(StatusPending) + "' "
	return r.listOrdersWithCount(ctx, filter, baseCond, []any{branchID}, BuildArchiveRoomOrderFilterClause, "o.updated_at DESC")
}

func (r *roomOrderRepository) ListAll(ctx context.Context, filter RoomOrderFilter) (*common.PaginatedResponse[[]*RoomOrderDTO], error) {
	return r.listOrders(ctx, filter, " ", []any{}, "o.created_at DESC")
}

func (r *roomOrderRepository) listOrders(ctx context.Context, filter RoomOrderFilter, baseCond string, baseArgs []any, orderBy string) (*common.PaginatedResponse[[]*RoomOrderDTO], error) {
	filterCond, args := BuildRoomOrderFilterClause(filter, baseArgs)
	offset := (filter.Page - 1) * filter.Limit
	argNum := len(args) + 1
	query := roomOrderWithRelations + baseCond + filterCond + " ORDER BY " + orderBy + " LIMIT $" + strconv.Itoa(argNum) + " OFFSET $" + strconv.Itoa(argNum+1)
	args = append(args, filter.Limit, offset)

	results, err := common.QueryRows(r.joinDAL, ctx, query, args, func(rows *sql.Rows) (*RoomOrderDTO, error) {
		var dto RoomOrderDTO
		if err := r.scanOrderRows(rows, &dto); err != nil {
			return nil, err
		}
		return &dto, nil
	})
	if err != nil {
		r.logger.Error("failed to list room orders", "error", err)
		return nil, common.ErrInternalServerError
	}
	return &common.PaginatedResponse[[]*RoomOrderDTO]{
		Data: results,
		Meta: common.BuildPaginationMeta(int64(len(results)), filter.Page, filter.Limit),
	}, nil
}

type roomOrderFilterClauseBuilder func(RoomOrderFilter, []any) (string, []any)

func (r *roomOrderRepository) listOrdersWithCount(
	ctx context.Context,
	filter RoomOrderFilter,
	baseCond string,
	baseArgs []any,
	buildClause roomOrderFilterClauseBuilder,
	orderBy string,
) (*common.PaginatedResponse[[]*RoomOrderDTO], error) {
	filterCond, args := buildClause(filter, baseArgs)

	countQuery := "SELECT COUNT(*) FROM room_orders o WHERE o.is_deleted = FALSE" + baseCond + filterCond
	var total int64
	err := r.joinDAL.QueryRow(ctx, countQuery, args, func(row *sql.Row) error {
		return row.Scan(&total)
	})
	if err != nil {
		r.logger.Error("failed to count room orders", "error", err)
		return nil, common.ErrInternalServerError
	}

	offset := (filter.Page - 1) * filter.Limit
	argNum := len(args) + 1
	query := roomOrderWithRelations + baseCond + filterCond + " ORDER BY " + orderBy + " LIMIT $" + strconv.Itoa(argNum) + " OFFSET $" + strconv.Itoa(argNum+1)
	listArgs := append(append([]any{}, args...), filter.Limit, offset)

	results, err := common.QueryRows(r.joinDAL, ctx, query, listArgs, func(rows *sql.Rows) (*RoomOrderDTO, error) {
		var dto RoomOrderDTO
		if err := r.scanOrderRows(rows, &dto); err != nil {
			return nil, err
		}
		return &dto, nil
	})
	if err != nil {
		r.logger.Error("failed to list room orders", "error", err)
		return nil, common.ErrInternalServerError
	}

	return &common.PaginatedResponse[[]*RoomOrderDTO]{
		Data: results,
		Meta: common.BuildPaginationMeta(total, filter.Page, filter.Limit),
	}, nil
}

func (r *roomOrderRepository) OrderNumberExists(ctx context.Context, branchID string, orderNumber int) (bool, error) {
	filter := map[string]any{
		"branch_id":    branchID,
		"order_number": orderNumber,
		"is_deleted":   false,
	}
	_, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		r.logger.Error("failed to check room order number", "error", err)
		return false, common.ErrInternalServerError
	}
	return true, nil
}

func (r *roomOrderRepository) SumActiveByBooking(ctx context.Context, bookingID string) (float64, error) {
	const query = `SELECT COALESCE(SUM(total), 0) FROM room_orders WHERE booking_id = $1 AND is_deleted = FALSE AND order_status <> $2`
	var total float64
	err := r.joinDAL.QueryRow(ctx, query, []any{bookingID, string(StatusCancelled)}, func(row *sql.Row) error {
		return row.Scan(&total)
	})
	if err != nil {
		r.logger.Error("failed to sum room orders for booking", "error", err)
		return 0, common.ErrInternalServerError
	}
	return total, nil
}

func (r *roomOrderRepository) scanOrderRow(row *sql.Row, dto *RoomOrderDTO) error {
	var orderItemsJSON []byte
	var cancellationReason sql.NullString
	err := row.Scan(
		&dto.ID, &dto.OrderNumber, &dto.RoomID, &dto.BookingID, &dto.BranchID, &dto.SessionKey, &dto.OrderStatus, &cancellationReason, &dto.Total, &dto.BillID,
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

func (r *roomOrderRepository) scanOrderRows(rows *sql.Rows, dto *RoomOrderDTO) error {
	var orderItemsJSON []byte
	var cancellationReason sql.NullString
	err := rows.Scan(
		&dto.ID, &dto.OrderNumber, &dto.RoomID, &dto.BookingID, &dto.BranchID, &dto.SessionKey, &dto.OrderStatus, &cancellationReason, &dto.Total, &dto.BillID,
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

func (r *roomOrderRepository) unmarshalOrderItems(dto *RoomOrderDTO, orderItemsJSON []byte) error {
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
