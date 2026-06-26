package category

import (
	"context"
	"database/sql"
	"errors"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"strings"
)

type CategoryRepository interface {
	Create(ctx context.Context, category Category) error
	Get(ctx context.Context, id string) (Category, error)
	Update(ctx context.Context, category Category) error
	Delete(ctx context.Context, id string) error
	UnDelete(ctx context.Context, id string) error
	List(ctx context.Context, filter common.Filter) (*common.PaginatedResponse[[]*Category], error)
	ListForBranch(ctx context.Context, branchID string) ([]*CategoryResponseSimplified, error)
	CheckExists(ctx context.Context, name string) error
	HardDeleteSoftDeletedByName(ctx context.Context, name string) error
}

type categoryRepository struct {
	dal    *common.DAL[*Category]
	logger config.Logger
}

func NewCategoryRepository(dal *common.DAL[*Category], logger config.Logger) CategoryRepository {
	return &categoryRepository{dal: dal, logger: logger}
}

func (r *categoryRepository) Create(ctx context.Context, category Category) error {
	_, err := r.dal.Create(ctx, &category)
	if err != nil {
		r.logger.Error("failed to create category", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *categoryRepository) Get(ctx context.Context, id string) (Category, error) {
	filter := map[string]any{"id": id}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("category not found", "error", err)
			return Category{}, common.ErrCategoryNotFound
		}
		r.logger.Error("failed to get category", "error", err)
		return Category{}, common.ErrInternalServerError
	}
	return *result, nil
}

func (r *categoryRepository) Update(ctx context.Context, category Category) error {
	filter := map[string]any{"id": category.ID}
	updates := map[string]any{
		"name":       category.Name,
		"icon":       category.Icon,
		"station":    category.Station,
		"deleted_at": category.DeletedAt,
		"is_deleted": category.IsDeleted,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("category not found", "error", err)
			return common.ErrCategoryNotFound
		}
		r.logger.Error("failed to update category", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *categoryRepository) Delete(ctx context.Context, id string) error {
	if _, err := r.Get(ctx, id); err != nil {
		return err
	}
	if err := r.dal.HardDelete(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common.ErrCategoryNotFound
		}
		r.logger.Error("failed to delete category", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *categoryRepository) List(ctx context.Context, filter common.Filter) (*common.PaginatedResponse[[]*Category], error) {
	common.NormalizeFilter(&filter)
	role, _ := middleware.GetRoleFromContext(ctx)

	filters := map[string]any{}
	if filter.Search != "" {
		filters["name"] = common.ILike(filter.Search)
	}

	offset := (filter.Page - 1) * filter.Limit
	var results []*Category
	var err error
	if role == "super_admin" {
		results, err = r.dal.ListIncludeDeleted(ctx, filters, filter.Limit, offset)
	} else {
		results, err = r.dal.List(ctx, filters, filter.Page, filter.Limit)
	}
	if err != nil {
		r.logger.Error("failed to list categories", "error", err)
		return nil, common.ErrInternalServerError
	}

	return &common.PaginatedResponse[[]*Category]{
		Data: results,
		Meta: common.BuildPaginationMeta(int64(len(results)), filter.Page, filter.Limit),
	}, nil	
}

func (r *categoryRepository) UnDelete(ctx context.Context, id string) error {
	filter := map[string]any{"id": id}
	updates := map[string]any{
		"is_deleted": false,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("category not found", "error", err)
			return common.ErrCategoryNotFound
		}
		r.logger.Error("failed to undelete category", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *categoryRepository) CheckExists(ctx context.Context, name string) error {
	lowerCaseName := strings.ToLower(name)
	filter := map[string]any{
		"name":       lowerCaseName,
		"is_deleted": false,
	}

	_, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		r.logger.Error("failed to check if category exists", "error", err)
		return common.ErrInternalServerError
	}

	return common.ErrCategoryAlreadyExists
}

func (r *categoryRepository) HardDeleteSoftDeletedByName(ctx context.Context, name string) error {
	if err := r.dal.HardDeleteByFilters(ctx, map[string]any{
		"name":       name,
		"is_deleted": true,
	}); err != nil {
		r.logger.Error("failed to purge soft-deleted category", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *categoryRepository) ListForBranch(ctx context.Context, branchID string) ([]*CategoryResponseSimplified, error) {
	filter := map[string]any{"branch_id": branchID}
	results, err := r.dal.List(ctx, filter, 1, 1000)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("no categories found for branch", "error", err)
			return nil, common.ErrNotFound
		}
		r.logger.Error("failed to list categories for branch", "error", err)
		return nil, common.ErrInternalServerError	
	}
	categoryDTOs := make([]*CategoryResponseSimplified, len(results))
	for i, category := range results {
		categoryDTOs[i] = category.ToResponseSimplified()
	}
	return categoryDTOs, nil
}