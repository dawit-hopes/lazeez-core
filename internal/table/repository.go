package table

import (
	"context"
	"database/sql"
	"errors"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type TableRepository interface {
	Create(ctx context.Context, table *Table) error
	Get(ctx context.Context, id string) (*Table, error)
	Update(ctx context.Context, table Table) error
	Delete(ctx context.Context, id string) error
	UnDelete(ctx context.Context, id string) error
	List(ctx context.Context, filters common.Filter, branchID string) (*common.PaginatedResponse[[]*Table], error)
	CheckExists(ctx context.Context, tableName, branchID string) error
}

type tableRepository struct {
	dal    *common.DAL[*Table]
	logger config.Logger
}

func NewTableRepository(dal *common.DAL[*Table], logger config.Logger) TableRepository {
	return &tableRepository{dal: dal, logger: logger}
}

func (r *tableRepository) Create(ctx context.Context, table *Table) error {
	_, err := r.dal.Create(ctx, table)
	if err != nil {
		return err
	}
	return nil
}

func (r *tableRepository) Get(ctx context.Context, id string) (*Table, error) {
	filter := map[string]any{"id": id}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("table not found", "error", err)
			return nil, common.ErrTableNotFound
		}
		r.logger.Error("failed to get table", "error", err)
		return nil, common.ErrInternalServerError
	}
	return result, nil
}

func (r *tableRepository) Update(ctx context.Context, table Table) error {
	updates := map[string]any{
		"table_name":      table.TableName,
		"branch_id":       table.BranchID,
		"reference":       table.Reference,
		"qr_code":         table.QRCode,
		"qr_version":      table.QRVersion,
		"status":          table.Status,
		"active_order_id": table.ActiveOrderID,
	}

	filter := map[string]any{"id": table.ID}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("table not found", "error", err)
			return common.ErrTableNotFound
		}
		r.logger.Error("failed to update table", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *tableRepository) Delete(ctx context.Context, id string) error {
	filter := map[string]any{"id": id}
	updates := map[string]any{"is_deleted": true}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("table not found", "error", err)
			return common.ErrTableNotFound
		}
		r.logger.Error("failed to delete table", "error", err)
		return common.ErrInternalServerError
	}
	return nil

}

func (r *tableRepository) UnDelete(ctx context.Context, id string) error {
	filter := map[string]any{"id": id}
	updates := map[string]any{"is_deleted": false}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("table not found", "error", err)
			return common.ErrTableNotFound
		}
		r.logger.Error("failed to undelete table", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *tableRepository) List(ctx context.Context, filters common.Filter, branchID string) (*common.PaginatedResponse[[]*Table], error) {
	common.NormalizeFilter(&filters)
	filter := map[string]any{"is_deleted": false}
	if branchID != "" {
		filter["branch_id"] = branchID
	}
	if filters.Search != "" {
		filter["table_name"] = common.ILike(filters.Search)
	}
	if filters.Filter != nil {
		for key, value := range filters.Filter {
			if key != "sort" && key != "order" {
				filter[key] = value
			}
		}
	}
	result, err := r.dal.List(ctx, filter, filters.Page, filters.Limit)
	if err != nil {
		r.logger.Error("failed to list tables", "error", err)
		return nil, common.ErrInternalServerError
	}

	meta := common.BuildPaginationMeta(int64(len(result)), filters.Page, filters.Limit)
	return &common.PaginatedResponse[[]*Table]{
		Data: result,
		Meta: meta,
	}, nil
}

func (r *tableRepository) CheckExists(ctx context.Context, tableName, branchID string) error {
	filter := map[string]any{"table_name": tableName, "branch_id": branchID, "is_deleted": false}
	_, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		r.logger.Error("failed to check if table exists", "error", err)
		return common.ErrInternalServerError
	}
	return common.ErrTableAlreadyExists
}
