package branch

import (
	"context"
	"database/sql"
	"errors"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
)

type BranchRepository interface {
	Create(ctx context.Context, branch Branch) error
	Get(ctx context.Context, id string) (Branch, error)
	Update(ctx context.Context, branch Branch) error
	Delete(ctx context.Context, id string) error
	CheckExists(ctx context.Context, merchantID, branchName, phoneNumber, excludeBranchID string) error
	List(ctx context.Context, filter common.Filter) (*common.PaginatedResponse[[]*Branch], error)
	ListByMerchantID(ctx context.Context, merchantID string, filter common.Filter) (*common.PaginatedResponse[[]*Branch], error)
	UnDelete(ctx context.Context, id string) error
}

type branchRepository struct {
	dal     *common.DAL[*Branch]
	joinDAL *common.JoinDAL
	logger  config.Logger
}

func NewBranchRepository(dal *common.DAL[*Branch], joinDAL *common.JoinDAL, logger config.Logger) BranchRepository {
	return &branchRepository{
		dal:     dal,
		joinDAL: joinDAL,
		logger:  logger,
	}
}

func (r *branchRepository) Create(ctx context.Context, branch Branch) error {
	_, err := r.dal.Create(ctx, &branch)
	if err != nil {
		r.logger.Error("failed to create branch", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *branchRepository) Get(ctx context.Context, id string) (Branch, error) {
	filter := map[string]any{"id": id}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("branch not found", "error", err)
			return Branch{}, common.ErrBranchNotFound
		}
		r.logger.Error("failed to get branch", "error", err)
		return Branch{}, common.ErrInternalServerError
	}
	return *result, nil
}

func (r *branchRepository) Update(ctx context.Context, branch Branch) error {
	filter := map[string]any{"id": branch.ID}
	updates := map[string]any{
		"merchant_id":  branch.MerchantID,
		"branch_name":  branch.BranchName,
		"address":      branch.Address,
		"phone_number": branch.PhoneNumber,
		"deleted_at":   branch.DeletedAt,
		"is_deleted":   branch.IsDeleted,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("branch not found", "error", err)
			return common.ErrBranchNotFound
		}
		r.logger.Error("failed to update branch", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *branchRepository) Delete(ctx context.Context, id string) error {
	const deleteBranchCascadeQuery = `
WITH updated_users AS (
	UPDATE users u
	SET is_deleted = TRUE, updated_at = NOW(), deleted_at = NOW()
	WHERE u.branch_id = $1 AND u.is_deleted = FALSE
)
UPDATE branches b
SET is_deleted = TRUE, updated_at = NOW(), deleted_at = NOW()
WHERE b.id = $1 AND b.is_deleted = FALSE;
`

	result, err := r.joinDAL.Exec(ctx, deleteBranchCascadeQuery, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("branch not found", "error", err)
			return common.ErrBranchNotFound
		}
		r.logger.Error("failed to cascade delete branch", "branch_id", id, "error", err)
		return common.ErrInternalServerError
	}

	affected, err := result.RowsAffected()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("branch not found", "error", err)
			return common.ErrBranchNotFound
		}
		r.logger.Error("failed to delete branch", "error", err)
		return common.ErrInternalServerError
	}
	if affected == 0 {
		r.logger.Error("branch not found", "branch_id", id)
		return common.ErrBranchNotFound
	}

	return nil
}

func (r *branchRepository) CheckExists(ctx context.Context, merchantID, branchName, phoneNumber, excludeBranchID string) error {
	query := `
		SELECT id FROM branches
		WHERE merchant_id = $1 AND is_deleted = FALSE
		AND (LOWER(branch_name) = LOWER($2) OR phone_number = $3)`
	args := []any{merchantID, branchName, phoneNumber}
	if excludeBranchID != "" {
		query += ` AND id != $4`
		args = append(args, excludeBranchID)
	}
	query += ` LIMIT 1`

	var id string
	err := r.joinDAL.QueryRow(ctx, query, args, func(row *sql.Row) error {
		return row.Scan(&id)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		r.logger.Error("failed to check if branch exists", "error", err)
		return common.ErrInternalServerError
	}
	return common.ErrBranchAlreadyExists
}

func (r *branchRepository) List(ctx context.Context, filter common.Filter) (*common.PaginatedResponse[[]*Branch], error) {
	role, _ := middleware.GetRoleFromContext(ctx)
	filters := map[string]any{}
	if filter.Search != "" {
		filters["branch_name"] = common.ILike(filter.Search)
	}
	offset := (filter.Page - 1) * filter.Limit
	var results []*Branch
	var err error
	if role == "super_admin" {
		results, err = r.dal.ListIncludeDeleted(ctx, filters, filter.Limit, offset)
	} else {
		results, err = r.dal.List(ctx, filters, filter.Page, filter.Limit)
	}
	if err != nil {
		r.logger.Error("failed to list branches", "error", err)
		return nil, common.ErrInternalServerError
	}
	return &common.PaginatedResponse[[]*Branch]{
		Data: results,
		Meta: common.BuildPaginationMeta(int64(len(results)), filter.Page, filter.Limit),
	}, nil
}

func (r *branchRepository) ListByMerchantID(ctx context.Context, merchantID string, filter common.Filter) (*common.PaginatedResponse[[]*Branch], error) {
	r.logger.Info("listing branches by merchant ID", "merchantID", merchantID)
	filters := map[string]any{"merchant_id": merchantID}
	if filter.Search != "" {
		filters["branch_name"] = common.ILike(filter.Search)
	}
	role, _ := middleware.GetRoleFromContext(ctx)
	offset := (filter.Page - 1) * filter.Limit
	var results []*Branch
	var err error
	if role == "super_admin" {
		results, err = r.dal.ListIncludeDeleted(ctx, filters, filter.Limit, offset)
	} else {
		results, err = r.dal.List(ctx, filters, filter.Page, filter.Limit)
	}
	if err != nil {
		r.logger.Error("failed to list branches by merchant ID", "error", err)
		return nil, common.ErrInternalServerError
	}
	return &common.PaginatedResponse[[]*Branch]{
		Data: results,
		Meta: common.BuildPaginationMeta(int64(len(results)), filter.Page, filter.Limit),
	}, nil
}

func (r *branchRepository) UnDelete(ctx context.Context, id string) error {
	filter := map[string]any{"id": id}
	updates := map[string]any{
		"is_deleted": false,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("branch not found", "error", err)
			return common.ErrBranchNotFound
		}
		r.logger.Error("failed to undelete branch", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}
