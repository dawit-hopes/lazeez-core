package rooms

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type ListScope string

const (
	ScopeMaster       ListScope = "master"
	ScopeBranchManage ListScope = "branch_manage"
)

type RoomRepository interface {
	CreateRoom(ctx context.Context, room Room) (*RoomDTO, error)
	GetRoom(ctx context.Context, id string) (*Room, error)
	GetMaster(ctx context.Context, id, merchantID string) (*Room, error)
	ListScoped(ctx context.Context, filter common.Filter, scope ListScope, branchID, merchantID string) (*common.PaginatedResponse[[]*RoomDTO], error)
	UpdateRoom(ctx context.Context, id string, room Room) error
	DeleteRoom(ctx context.Context, id string) error
	CheckMasterExists(ctx context.Context, name, merchantID string) error
	CheckBranchExists(ctx context.Context, name, branchID string) error
	CheckCloneExists(ctx context.Context, branchID, parentID string) error
}

type roomRepository struct {
	dal    *common.DAL[*Room]
	join   *common.JoinDAL
	logger config.Logger
}

func NewRoomRepository(dal *common.DAL[*Room], join *common.JoinDAL, logger config.Logger) RoomRepository {
	return &roomRepository{dal: dal, join: join, logger: logger}
}

func (r *roomRepository) CreateRoom(ctx context.Context, room Room) (*RoomDTO, error) {
	_, err := r.dal.Create(ctx, &room)
	if err != nil {
		r.logger.Error("failed to create room", "error", err)
		return nil, common.ErrInternalServerError
	}
	dto := room.ToDTO()
	return &dto, nil
}

func (r *roomRepository) GetRoom(ctx context.Context, id string) (*Room, error) {
	filter := map[string]any{"id": id}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrRoomNotFound
		}
		r.logger.Error("failed to get room", "error", err)
		return nil, common.ErrInternalServerError
	}
	return result, nil
}

func (r *roomRepository) GetMaster(ctx context.Context, id, merchantID string) (*Room, error) {
	filter := map[string]any{
		"id":         id,
		"branch_id":  common.IsNull{},
		"parent_id":  common.IsNull{},
		"is_deleted": false,
	}
	if merchantID != "" {
		filter["merchant_id"] = merchantID
	}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrRoomNotFound
		}
		r.logger.Error("failed to get master room", "error", err)
		return nil, common.ErrInternalServerError
	}
	return result, nil
}

func (r *roomRepository) ListScoped(ctx context.Context, filter common.Filter, scope ListScope, branchID, merchantID string) (*common.PaginatedResponse[[]*RoomDTO], error) {
	switch scope {
	case ScopeMaster:
		return r.listMaster(ctx, filter, merchantID)
	case ScopeBranchManage:
		if branchID == "" {
			return nil, common.ErrUnAuthorized
		}
		return r.listBranchManage(ctx, filter, branchID)
	default:
		return nil, common.ErrUnAuthorized
	}
}

func (r *roomRepository) listMaster(ctx context.Context, filter common.Filter, merchantID string) (*common.PaginatedResponse[[]*RoomDTO], error) {
	if merchantID == "" {
		return nil, common.ErrUnAuthorized
	}
	filters := map[string]any{
		"merchant_id": merchantID,
		"branch_id":   common.IsNull{},
		"parent_id":   common.IsNull{},
	}
	if filter.Search != "" {
		filters["name"] = common.ILike(filter.Search)
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	results, err := r.dal.List(ctx, filters, page, limit)
	if err != nil {
		r.logger.Error("failed to list master rooms", "error", err)
		return nil, common.ErrInternalServerError
	}
	dtos := make([]*RoomDTO, len(results))
	for i, room := range results {
		dto := room.ToDTO()
		dtos[i] = &dto
	}
	return &common.PaginatedResponse[[]*RoomDTO]{
		Data: dtos,
		Meta: common.BuildPaginationMeta(int64(len(dtos)), page, limit),
	}, nil
}

func (r *roomRepository) listBranchManage(ctx context.Context, filter common.Filter, branchID string) (*common.PaginatedResponse[[]*RoomDTO], error) {
	searchClause := ""
	args := []any{branchID}
	if filter.Search != "" {
		searchClause = " AND r.name ILIKE $2 ESCAPE '\\'"
		args = append(args, "%"+filter.Search+"%")
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit
	argN := len(args)

	query := fmt.Sprintf(`
		SELECT r.id, r.name, r.description, r.merchant_id, r.branch_id, r.parent_id,
			r.price_per_night, r.deleted_at, r.is_deleted, r.created_at, r.updated_at,
			(r.branch_id IS NULL AND r.parent_id IS NULL) AS is_master,
			(r.parent_id IS NOT NULL) AS is_clone,
			(r.branch_id IS NOT NULL) AS can_edit
		FROM rooms r
		INNER JOIN branches b ON b.id = $1 AND b.is_deleted = FALSE
		WHERE r.is_deleted = FALSE
			AND (
				(r.branch_id IS NULL AND r.parent_id IS NULL AND r.merchant_id = b.merchant_id
					AND NOT EXISTS (
						SELECT 1 FROM rooms c
						WHERE c.parent_id = r.id AND c.branch_id = b.id AND c.is_deleted = FALSE
					))
				OR r.branch_id = b.id
			)
			%s
		ORDER BY is_master DESC, r.created_at DESC
		LIMIT $%d OFFSET $%d`, searchClause, argN+1, argN+2)

	args = append(args, limit, offset)

	rows, err := common.QueryRows(r.join, ctx, query, args, scanRoomListRow)
	if err != nil {
		r.logger.Error("failed to list branch rooms", "error", err)
		return nil, common.ErrInternalServerError
	}

	dtos := make([]*RoomDTO, len(rows))
	for i := range rows {
		dtos[i] = &rows[i]
	}
	return &common.PaginatedResponse[[]*RoomDTO]{
		Data: dtos,
		Meta: common.BuildPaginationMeta(int64(len(dtos)), page, limit),
	}, nil
}

func scanRoomListRow(rows *sql.Rows) (RoomDTO, error) {
	var room Room
	var isMaster, isClone, canEdit bool
	if err := rows.Scan(
		&room.ID, &room.Name, &room.Description, &room.MerchantID, &room.BranchID, &room.ParentID,
		&room.PricePerNight, &room.DeletedAt, &room.IsDeleted, &room.CreatedAt, &room.UpdatedAt,
		&isMaster, &isClone, &canEdit,
	); err != nil {
		return RoomDTO{}, err
	}
	dto := room.ToDTO()
	dto.IsMaster = isMaster
	dto.IsClone = isClone
	dto.CanEdit = canEdit
	return dto, nil
}

func (r *roomRepository) UpdateRoom(ctx context.Context, id string, room Room) error {
	filter := map[string]any{"id": id}
	updates := map[string]any{
		"name":            room.Name,
		"description":     room.Description,
		"price_per_night": room.PricePerNight,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common.ErrRoomNotFound
		}
		r.logger.Error("failed to update room", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *roomRepository) DeleteRoom(ctx context.Context, id string) error {
	filter := map[string]any{"id": id}
	updates := map[string]any{"is_deleted": true}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common.ErrRoomNotFound
		}
		r.logger.Error("failed to delete room", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *roomRepository) CheckMasterExists(ctx context.Context, name, merchantID string) error {
	filter := map[string]any{
		"name":        name,
		"merchant_id": merchantID,
		"branch_id":   common.IsNull{},
		"parent_id":   common.IsNull{},
		"is_deleted":  false,
	}
	_, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		r.logger.Error("failed to check master room exists", "error", err)
		return common.ErrInternalServerError
	}
	return common.ErrRoomAlreadyExists
}

func (r *roomRepository) CheckBranchExists(ctx context.Context, name, branchID string) error {
	filter := map[string]any{
		"name":       name,
		"branch_id":  branchID,
		"parent_id":  common.IsNull{},
		"is_deleted": false,
	}
	_, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		r.logger.Error("failed to check branch room exists", "error", err)
		return common.ErrInternalServerError
	}
	return common.ErrRoomAlreadyExists
}

func (r *roomRepository) CheckCloneExists(ctx context.Context, branchID, parentID string) error {
	filter := map[string]any{
		"branch_id":  branchID,
		"parent_id":  parentID,
		"is_deleted": false,
	}
	_, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		r.logger.Error("failed to check room clone exists", "error", err)
		return common.ErrInternalServerError
	}
	return common.ErrRoomCloneAlreadyExists
}
