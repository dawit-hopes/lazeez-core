package booking

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type BookingRepository interface {
	Create(ctx context.Context, b *Booking) error
	Get(ctx context.Context, id string) (*Booking, error)
	Update(ctx context.Context, b Booking) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter common.Filter, branchID, merchantID string) (*common.PaginatedResponse[[]*Booking], error)
	HasActiveBooking(ctx context.Context, roomID string) (bool, error)
	GetActiveByRoomID(ctx context.Context, roomID string) (*Booking, error)
	SetPasscodeAttempts(ctx context.Context, id string, attempts int, lockedUntil sql.NullTime) error
}

type bookingRepository struct {
	dal    *common.DAL[*Booking]
	join   *common.JoinDAL
	logger config.Logger
}

func NewBookingRepository(dal *common.DAL[*Booking], join *common.JoinDAL, logger config.Logger) BookingRepository {
	return &bookingRepository{dal: dal, join: join, logger: logger}
}

func (r *bookingRepository) Create(ctx context.Context, b *Booking) error {
	if _, err := r.dal.Create(ctx, b); err != nil {
		r.logger.Error("failed to create booking", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *bookingRepository) Get(ctx context.Context, id string) (*Booking, error) {
	filter := map[string]any{"id": id, "is_deleted": false}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrBookingNotFound
		}
		r.logger.Error("failed to get booking", "error", err)
		return nil, common.ErrInternalServerError
	}
	return result, nil
}

func (r *bookingRepository) Update(ctx context.Context, b Booking) error {
	filter := map[string]any{"id": b.ID, "is_deleted": false}
	updates := map[string]any{
		"guest_name":       b.GuestName,
		"guest_phone":      b.GuestPhone,
		"number_of_nights": b.NumberOfNights,
		"check_in_date":    b.CheckInDate,
		"check_out_date":   b.CheckOutDate,
		"status":           b.Status,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common.ErrBookingNotFound
		}
		r.logger.Error("failed to update booking", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *bookingRepository) Delete(ctx context.Context, id string) error {
	if err := r.dal.Delete(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common.ErrBookingNotFound
		}
		r.logger.Error("failed to delete booking", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

// GetActiveByRoomID returns the active booking for a room, or ErrRoomNotOccupied
// when the room currently has no active booking.
func (r *bookingRepository) GetActiveByRoomID(ctx context.Context, roomID string) (*Booking, error) {
	filter := map[string]any{
		"room_id":    roomID,
		"status":     string(BookingActive),
		"is_deleted": false,
	}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrRoomNotOccupied
		}
		r.logger.Error("failed to get active booking by room", "error", err)
		return nil, common.ErrInternalServerError
	}
	return result, nil
}

// SetPasscodeAttempts records the failed-attempt counter and optional lockout
// timestamp for a booking's passcode verification.
func (r *bookingRepository) SetPasscodeAttempts(ctx context.Context, id string, attempts int, lockedUntil sql.NullTime) error {
	filter := map[string]any{"id": id, "is_deleted": false}
	updates := map[string]any{
		"passcode_attempts":     attempts,
		"passcode_locked_until": lockedUntil,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common.ErrBookingNotFound
		}
		r.logger.Error("failed to update passcode attempts", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *bookingRepository) HasActiveBooking(ctx context.Context, roomID string) (bool, error) {
	filter := map[string]any{
		"room_id":    roomID,
		"status":     string(BookingActive),
		"is_deleted": false,
	}
	_, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		r.logger.Error("failed to check active booking", "error", err)
		return false, common.ErrInternalServerError
	}
	return true, nil
}

func (r *bookingRepository) List(ctx context.Context, filter common.Filter, branchID, merchantID string) (*common.PaginatedResponse[[]*Booking], error) {
	common.NormalizeFilter(&filter)
	if merchantID != "" && branchID == "" {
		return r.listByMerchant(ctx, filter, merchantID)
	}

	filters := map[string]any{"is_deleted": false}
	if branchID != "" {
		filters["branch_id"] = branchID
	}
	if status := filterStatus(filter); status != "" {
		filters["status"] = status
	}
	if filter.Search != "" {
		filters["guest_name"] = common.ILike(filter.Search)
	}

	page, limit := filter.PageLimit()

	total, err := r.dal.CountFiltered(ctx, filters)
	if err != nil {
		r.logger.Error("failed to count bookings", "error", err)
		return nil, common.ErrInternalServerError
	}
	results, err := r.dal.List(ctx, filters, page, limit)
	if err != nil {
		r.logger.Error("failed to list bookings", "error", err)
		return nil, common.ErrInternalServerError
	}
	return &common.PaginatedResponse[[]*Booking]{
		Data: results,
		Meta: common.BuildPaginationMeta(total, page, limit),
	}, nil
}

func (r *bookingRepository) listByMerchant(ctx context.Context, filter common.Filter, merchantID string) (*common.PaginatedResponse[[]*Booking], error) {
	common.NormalizeFilter(&filter)
	args := []any{merchantID}
	extraClause := ""
	if status := filterStatus(filter); status != "" {
		args = append(args, status)
		extraClause += fmt.Sprintf(" AND bk.status = $%d", len(args))
	}
	if filter.Search != "" {
		args = append(args, common.ILikePattern(filter.Search))
		extraClause += fmt.Sprintf(" AND bk.guest_name ILIKE $%d ESCAPE '\\'", len(args))
	}

	page, limit := filter.PageLimit()
	offset := (page - 1) * limit

	whereClause := fmt.Sprintf(`bk.is_deleted = FALSE AND br.merchant_id = $1%s`, extraClause)

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM bookings bk
		INNER JOIN branches br ON br.id = bk.branch_id AND br.is_deleted = FALSE
		WHERE %s`, whereClause)

	var total int64
	if err := r.join.QueryRow(ctx, countQuery, args, func(row *sql.Row) error {
		return row.Scan(&total)
	}); err != nil {
		r.logger.Error("failed to count merchant bookings", "error", err)
		return nil, common.ErrInternalServerError
	}

	argN := len(args)
	listQuery := fmt.Sprintf(`
		SELECT bk.id, bk.room_id, bk.branch_id, bk.guest_name, bk.guest_phone,
			bk.number_of_nights, bk.check_in_date, bk.check_out_date, bk.passcode_hash,
			bk.passcode_attempts, bk.passcode_locked_until,
			bk.status, bk.deleted_at, bk.is_deleted, bk.created_at, bk.updated_at
		FROM bookings bk
		INNER JOIN branches br ON br.id = bk.branch_id AND br.is_deleted = FALSE
		WHERE %s
		ORDER BY bk.created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argN+1, argN+2)

	listArgs := append(append([]any{}, args...), limit, offset)

	rows, err := common.QueryRows(r.join, ctx, listQuery, listArgs, func(rows *sql.Rows) (*Booking, error) {
		var b Booking
		if err := rows.Scan(b.Addr()...); err != nil {
			return nil, err
		}
		return &b, nil
	})
	if err != nil {
		r.logger.Error("failed to list merchant bookings", "error", err)
		return nil, common.ErrInternalServerError
	}

	return &common.PaginatedResponse[[]*Booking]{
		Data: rows,
		Meta: common.BuildPaginationMeta(total, page, limit),
	}, nil
}

func filterStatus(filter common.Filter) string {
	if filter.Filter == nil {
		return ""
	}
	if v, ok := filter.Filter["status"].(string); ok {
		return v
	}
	return ""
}
