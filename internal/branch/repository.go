package branch

import (
	"context"
	"database/sql"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
)

type BranchRepository interface {
	Create(ctx context.Context, branch Branch) error
	Get(ctx context.Context, id string) (Branch, error)
	Update(ctx context.Context, branch Branch) error
	Delete(ctx context.Context, id string) error
	CheckExists(ctx context.Context, merchantID string, branchName, phoneNumber string) error
	GetAll(ctx context.Context) ([]*Branch, error)
	GetAllByMerchantID(ctx context.Context, merchantID string) ([]*Branch, error)
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
		if err == sql.ErrNoRows {
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
		if err == sql.ErrNoRows {
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
		r.logger.Error("failed to cascade delete branch", "branch_id", id, "error", err)
		return common.ErrInternalServerError
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		r.logger.Error("branch not found", "branch_id", id)
		return common.ErrBranchNotFound
	}

	return nil
}

func (r *branchRepository) CheckExists(ctx context.Context, merchantID string, branchName, phoneNumber string) error {
	filter := map[string]any{"merchant_id": merchantID, "branch_name": branchName, "phone_number": phoneNumber}
	_, err := r.dal.Get(ctx, filter)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		r.logger.Error("failed to check if branch exists", "error", err)
		return common.ErrInternalServerError
	}
	return common.ErrBranchAlreadyExists
}

func (r *branchRepository) GetAll(ctx context.Context) ([]*Branch, error) {
	role, _ := middleware.GetRoleFromContext(ctx)

	var results []*Branch
	var err error
	if role == "super_admin" {
		results, err = r.dal.ListIncludeDeleted(ctx, map[string]any{}, 0, 0)
	} else {
		results, err = r.dal.List(ctx, map[string]any{}, 0, 0)
	}
	if err != nil {
		r.logger.Error("failed to get all branches", "error", err)
		return nil, err
	}
	return results, nil
}

func (r *branchRepository) GetAllByMerchantID(ctx context.Context, merchantID string) ([]*Branch, error) {
	r.logger.Info("getting all branches by merchant ID", "merchantID", merchantID)
	filter := map[string]any{"merchant_id": merchantID}

	role, _ := middleware.GetRoleFromContext(ctx)

	var results []*Branch
	var err error
	if role == "super_admin" {
		results, err = r.dal.ListIncludeDeleted(ctx, filter, 0, 0)
	} else {
		results, err = r.dal.List(ctx, filter, 0, 0)
	}
	if err != nil {
		r.logger.Error("failed to get all branches by merchant ID", "error", err)
		return nil, err
	}
	return results, nil
}
