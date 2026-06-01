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
	UpdateStatus(ctx context.Context, id, status string) error
	Get(ctx context.Context, id string) (*RoomWithType, error)
	GetByReference(ctx context.Context, reference string) (*Room, error)
	Update(ctx context.Context, room Room) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter common.Filter, branchID, merchantID string) (*common.PaginatedResponse[[]*RoomWithType], error)
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

func (r *roomRepository) UpdateStatus(ctx context.Context, id, status string) error {
	filter := map[string]any{"id": id, "is_deleted": false}
	updates := map[string]any{"status": status}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common.ErrRoomNotFound
		}
		r.logger.Error("failed to update single room status", "room_id", id, "status", status, "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *roomRepository) Get(ctx context.Context, id string) (*RoomWithType, error) {
	query := roomWithTypeSelect + ` WHERE sr.id = $1 AND sr.is_deleted = FALSE`
	var result *RoomWithType
	err := r.join.QueryRow(ctx, query, []any{id}, func(row *sql.Row) error {
		var err error
		result, err = scanRoomWithType(row)
		return err
	})
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

func (r *roomRepository) List(ctx context.Context, filter common.Filter, branchID, merchantID string) (*common.PaginatedResponse[[]*RoomWithType], error) {
	if merchantID != "" && branchID == "" {
		return r.listByMerchant(ctx, filter, merchantID)
	}
	return r.listByBranch(ctx, filter, branchID)
}

func (r *roomRepository) listByBranch(ctx context.Context, filter common.Filter, branchID string) (*common.PaginatedResponse[[]*RoomWithType], error) {
	common.NormalizeFilter(&filter)
	args := []any{}
	conditions := []string{"sr.is_deleted = FALSE"}
	argN := 1

	if branchID != "" {
		conditions = append(conditions, fmt.Sprintf("sr.branch_id = $%d", argN))
		args = append(args, branchID)
		argN++
	}
	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("sr.room_number ILIKE $%d ESCAPE '\\'", argN))
		args = append(args, common.ILikePattern(filter.Search))
		argN++
	}

	page, limit := filter.PageLimit()
	offset := (page - 1) * limit

	whereClause := "WHERE " + joinConditions(conditions)

	countQuery := `SELECT COUNT(*) FROM single_rooms sr ` + whereClause
	var total int64
	if err := r.join.QueryRow(ctx, countQuery, args, func(row *sql.Row) error {
		return row.Scan(&total)
	}); err != nil {
		r.logger.Error("failed to count single rooms", "error", err)
		return nil, common.ErrInternalServerError
	}

	listQuery := roomWithTypeSelect + " " + whereClause +
		fmt.Sprintf(" ORDER BY sr.created_at DESC LIMIT $%d OFFSET $%d", argN, argN+1)
	listArgs := append(append([]any{}, args...), limit, offset)

	rows, err := common.QueryRows(r.join, ctx, listQuery, listArgs, func(rows *sql.Rows) (*RoomWithType, error) {
		return scanRoomWithType(rows)
	})
	if err != nil {
		r.logger.Error("failed to list single rooms", "error", err)
		return nil, common.ErrInternalServerError
	}

	return &common.PaginatedResponse[[]*RoomWithType]{
		Data: rows,
		Meta: common.BuildPaginationMeta(total, page, limit),
	}, nil
}

func (r *roomRepository) listByMerchant(ctx context.Context, filter common.Filter, merchantID string) (*common.PaginatedResponse[[]*RoomWithType], error) {
	common.NormalizeFilter(&filter)
	args := []any{merchantID}
	searchClause := ""
	if filter.Search != "" {
		searchClause = " AND sr.room_number ILIKE $2 ESCAPE '\\'"
		args = append(args, common.ILikePattern(filter.Search))
	}

	page, limit := filter.PageLimit()
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
	listQuery := roomWithTypeSelect + `
		INNER JOIN branches b ON b.id = sr.branch_id AND b.is_deleted = FALSE
		WHERE ` + whereClause + fmt.Sprintf(`
		ORDER BY sr.created_at DESC
		LIMIT $%d OFFSET $%d`, argN+1, argN+2)

	listArgs := append(append([]any{}, args...), limit, offset)

	rows, err := common.QueryRows(r.join, ctx, listQuery, listArgs, func(rows *sql.Rows) (*RoomWithType, error) {
		return scanRoomWithType(rows)
	})
	if err != nil {
		r.logger.Error("failed to list merchant single rooms", "error", err)
		return nil, common.ErrInternalServerError
	}

	return &common.PaginatedResponse[[]*RoomWithType]{
		Data: rows,
		Meta: common.BuildPaginationMeta(total, page, limit),
	}, nil
}

func joinConditions(conditions []string) string {
	if len(conditions) == 0 {
		return ""
	}
	out := conditions[0]
	for i := 1; i < len(conditions); i++ {
		out += " AND " + conditions[i]
	}
	return out
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
