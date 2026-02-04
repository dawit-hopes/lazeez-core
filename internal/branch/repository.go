package branch

import (
	"context"
	"database/sql"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type BranchRepository interface {
	Create(ctx context.Context, branch Branch) (Branch, error)
	Get(ctx context.Context, id string) (Branch, error)
	Update(ctx context.Context, branch Branch) (Branch, error)
	Delete(ctx context.Context, id string) error
}

type branchRepository struct {
	dal    *common.DAL[*Branch]
	logger config.Logger
}

func NewBranchRepository(dal *common.DAL[*Branch], logger config.Logger) BranchRepository {
	return &branchRepository{dal: dal, logger: logger}
}

func (r *branchRepository) Create(ctx context.Context, branch Branch) (Branch, error) {
	result, err := r.dal.Create(ctx, &branch)
	if err != nil {
		r.logger.Error("failed to create branch", "error", err)
		return branch, err
	}
	return *result, nil
}

func (r *branchRepository) Get(ctx context.Context, id string) (Branch, error) {
	filter := map[string]any{"id": id}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Error("branch not found", "error", err)
			return Branch{}, err
		}
		r.logger.Error("failed to get branch", "error", err)
		return Branch{}, err
	}
	return *result, nil
}

func (r *branchRepository) Update(ctx context.Context, branch Branch) (Branch, error) {
	result, err := r.dal.Update(ctx, branch.ID, &branch)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Error("branch not found", "error", err)
			return Branch{}, err
		}
		r.logger.Error("failed to update branch", "error", err)
		return Branch{}, err
	}
	return *result, nil
}

func (r *branchRepository) Delete(ctx context.Context, id string) error {
	err := r.dal.Delete(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Error("branch not found", "error", err)
			return err
		}
		r.logger.Error("failed to delete branch", "error", err)
		return err
	}
	return nil
}
