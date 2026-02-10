package branch

import (
	"context"
	"database/sql"
	"lazeez-core/config"
	"lazeez-core/internal/common"
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
	dal    *common.DAL[*Branch]
	logger config.Logger
}

func NewBranchRepository(dal *common.DAL[*Branch], logger config.Logger) BranchRepository {
	return &branchRepository{dal: dal, logger: logger}
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
	_, err := r.dal.Update(ctx, branch.ID, &branch)
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
	err := r.dal.Delete(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Error("branch not found", "error", err)
			return common.ErrBranchNotFound
		}
		r.logger.Error("failed to delete branch", "error", err)
		return common.ErrInternalServerError
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
	results, err := r.dal.List(ctx, map[string]any{}, 0, 0)
	if err != nil {
		r.logger.Error("failed to get all branches", "error", err)
		return nil, err
	}
	return results, nil
}

func (r *branchRepository) GetAllByMerchantID(ctx context.Context, merchantID string) ([]*Branch, error) {
	r.logger.Info("getting all branches by merchant ID", "merchantID", merchantID)
	filter := map[string]any{"merchant_id": merchantID}
	results, err := r.dal.List(ctx, filter, 0, 0)
	if err != nil {
		r.logger.Error("failed to get all branches by merchant ID", "error", err)
		return nil, err
	}
	return results, nil
}
