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
	List(ctx context.Context) ([]*Category, error)
	CheckExists(ctx context.Context, name string) error
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
	filter := map[string]any{"id": id}
	updates := map[string]any{
		"is_deleted": true,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("category not found", "error", err)
			return common.ErrCategoryNotFound
		}
		r.logger.Error("failed to delete category", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *categoryRepository) List(ctx context.Context) ([]*Category, error) {
	role, _ := middleware.GetRoleFromContext(ctx)

	var results []*Category
	var err error
	if role == "super_admin" {
		results, err = r.dal.ListIncludeDeleted(ctx, map[string]any{}, 0, 0)
	} else {
		results, err = r.dal.List(ctx, map[string]any{}, 0, 0)
	}
	if err != nil {
		r.logger.Error("failed to list categories", "error", err)
		return nil, common.ErrInternalServerError
	}

	return results, nil
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
