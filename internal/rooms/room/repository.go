package room

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type RoomRepository interface {
	Create(ctx context.Context, room *Room) error
	Get(ctx context.Context, id string) (*Room, error)
	GetByReference(ctx context.Context, reference string) (*Room, error)
	Update(ctx context.Context, room Room) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter common.Filter, branchID, merchantID string) (*common.PaginatedResponse[[]*Room], error)
	CheckRoomNumberExists(ctx context.Context, branchID, roomNumber string) error
}

type roomRepository struct {
	dal    *common.DAL[*Room]
	join   *common.JoinDAL
	logger config.Logger
}

func NewRoomRepository(dal *common.DAL[*Room], join *common.JoinDAL, logger config.Logger) RoomRepository {
	return &roomRepository{dal: dal, join: join, logger: logger}
}

func (r *roomRepository) Create(ctx context.Context, room *Room) error {
	if _, err := r.dal.Create(ctx, room); err != nil {
		r.logger.Error("failed to create single room", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *roomRepository) Get(ctx context.Context, id string) (*Room, error) {
	filter := map[string]any{"id": id, "is_deleted": false}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrRoomNotFound
		}
		r.logger.Error("failed to get single room", "error", err)
		return nil, common.ErrInternalServerError
	}
	return result, nil
}

// GetByReference resolves a physical room from its permanent QR reference token.
func (r *roomRepository) GetByReference(ctx context.Context, reference string) (*Room, error) {
	filter := map[string]any{"reference": reference, "is_deleted": false}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrReferenceNotValid
		}
		r.logger.Error("failed to get single room by reference", "error", err)
		return nil, common.ErrInternalServerError
	}
	return result, nil
}

func (r *roomRepository) Update(ctx context.Context, room Room) error {
	filter := map[string]any{"id": room.ID, "is_deleted": false}
	updates := map[string]any{
		"room_number":  room.RoomNumber,
		"room_type_id": room.RoomTypeID,
		"floor":        room.Floor,
		"qr_code":      room.QRCode,
		"qr_version":   room.QRVersion,
		"reference":    room.Reference,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common.ErrRoomNotFound
		}
		r.logger.Error("failed to update single room", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *roomRepository) Delete(ctx context.Context, id string) error {
	if err := r.dal.HardDelete(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common.ErrRoomNotFound
		}
		r.logger.Error("failed to delete single room", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *roomRepository) List(ctx context.Context, filter common.Filter, branchID, merchantID string) (*common.PaginatedResponse[[]*Room], error) {
	if merchantID != "" && branchID == "" {
		return r.listByMerchant(ctx, filter, merchantID)
	}

	filters := map[string]any{"is_deleted": false}
	if branchID != "" {
		filters["branch_id"] = branchID
	}
	if filter.Search != "" {
		filters["room_number"] = common.ILike(filter.Search)
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}

	total, err := r.dal.CountFiltered(ctx, filters)
	if err != nil {
		r.logger.Error("failed to count single rooms", "error", err)
		return nil, common.ErrInternalServerError
	}
	results, err := r.dal.List(ctx, filters, page, limit)
	if err != nil {
		r.logger.Error("failed to list single rooms", "error", err)
		return nil, common.ErrInternalServerError
	}
	return &common.PaginatedResponse[[]*Room]{
		Data: results,
		Meta: common.BuildPaginationMeta(total, page, limit),
	}, nil
}

func (r *roomRepository) listByMerchant(ctx context.Context, filter common.Filter, merchantID string) (*common.PaginatedResponse[[]*Room], error) {
	args := []any{merchantID}
	searchClause := ""
	if filter.Search != "" {
		searchClause = " AND sr.room_number ILIKE $2 ESCAPE '\\'"
		args = append(args, common.ILikePattern(filter.Search))
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit

	whereClause := fmt.Sprintf(`sr.is_deleted = FALSE AND b.merchant_id = $1%s`, searchClause)

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM single_rooms sr
		INNER JOIN branches b ON b.id = sr.branch_id AND b.is_deleted = FALSE
		WHERE %s`, whereClause)

	var total int64
	if err := r.join.QueryRow(ctx, countQuery, args, func(row *sql.Row) error {
		return row.Scan(&total)
	}); err != nil {
		r.logger.Error("failed to count merchant single rooms", "error", err)
		return nil, common.ErrInternalServerError
	}

	argN := len(args)
	listQuery := fmt.Sprintf(`
		SELECT sr.id, sr.room_number, sr.room_type_id, sr.branch_id, sr.floor,
			sr.qr_code, sr.qr_version, sr.reference, sr.deleted_at, sr.is_deleted,
			sr.created_at, sr.updated_at
		FROM single_rooms sr
		INNER JOIN branches b ON b.id = sr.branch_id AND b.is_deleted = FALSE
		WHERE %s
		ORDER BY sr.created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argN+1, argN+2)

	listArgs := append(append([]any{}, args...), limit, offset)

	rows, err := common.QueryRows(r.join, ctx, listQuery, listArgs, func(rows *sql.Rows) (*Room, error) {
		var room Room
		if err := rows.Scan(room.Addr()...); err != nil {
			return nil, err
		}
		return &room, nil
	})
	if err != nil {
		r.logger.Error("failed to list merchant single rooms", "error", err)
		return nil, common.ErrInternalServerError
	}

	return &common.PaginatedResponse[[]*Room]{
		Data: rows,
		Meta: common.BuildPaginationMeta(total, page, limit),
	}, nil
}

func (r *roomRepository) CheckRoomNumberExists(ctx context.Context, branchID, roomNumber string) error {
	filter := map[string]any{
		"branch_id":   branchID,
		"room_number": roomNumber,
		"is_deleted":  false,
	}
	_, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		r.logger.Error("failed to check single room number exists", "error", err)
		return common.ErrInternalServerError
	}
	return common.ErrRoomNumberExists
}
