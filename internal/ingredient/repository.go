package ingredient

import (
	"context"
	"database/sql"
	"errors"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type IngredientRepository interface {
	Create(ctx context.Context, ingredient Ingredient) error
	Get(ctx context.Context, id string) (Ingredient, error)
	Update(ctx context.Context, ingredient Ingredient) error
	List(ctx context.Context, filter common.Filter) (*common.PaginatedResponse[[]*Ingredient], error)
	CheckExists(ctx context.Context, name string) error
	Delete(ctx context.Context, id string) error
	UnDelete(ctx context.Context, id string) error
}

type ingredientRepository struct {
	dal    *common.DAL[*Ingredient]
	logger config.Logger
}

func NewIngredientRepository(dal *common.DAL[*Ingredient], logger config.Logger) IngredientRepository {
	return &ingredientRepository{dal: dal, logger: logger}
}

func (r *ingredientRepository) Create(ctx context.Context, ingredient Ingredient) error {
	_, err := r.dal.Create(ctx, &ingredient)
	if err != nil {
		r.logger.Error("failed to create ingredient", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *ingredientRepository) Get(ctx context.Context, id string) (Ingredient, error) {
	filter := map[string]any{"id": id}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("ingredient not found", "error", err)
			return Ingredient{}, common.ErrIngredientNotFound
		}
		r.logger.Error("failed to get ingredient", "error", err)
		return Ingredient{}, common.ErrInternalServerError
	}
	return *result, nil
}

func (r *ingredientRepository) Update(ctx context.Context, ingredient Ingredient) error {
	filter := map[string]any{"id": ingredient.ID}
	updates := map[string]any{
		"name":       ingredient.Name,
		"deleted_at": ingredient.DeletedAt,
		"is_deleted": ingredient.IsDeleted,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("ingredient not found", "error", err)
			return common.ErrIngredientNotFound
		}
		r.logger.Error("failed to update ingredient", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *ingredientRepository) List(ctx context.Context, filter common.Filter) (*common.PaginatedResponse[[]*Ingredient], error) {
	common.NormalizeFilter(&filter)
	filters := map[string]any{}
	if filter.Search != "" {
		filters["name"] = common.ILike(filter.Search)
	}
	results, err := r.dal.List(ctx, filters, filter.Page, filter.Limit)
	if err != nil {
		r.logger.Error("failed to list ingredients", "error", err)
		return nil, common.ErrInternalServerError
	}
	return &common.PaginatedResponse[[]*Ingredient]{
		Data: results,
		Meta: common.BuildPaginationMeta(int64(len(results)), filter.Page, filter.Limit),
	}, nil
}

func (r *ingredientRepository) CheckExists(ctx context.Context, name string) error {
	filter := map[string]any{"name": name, "is_deleted": false}
	_, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		r.logger.Error("failed to check if ingredient exists", "error", err)
		return common.ErrInternalServerError
	}
	return common.ErrIngredientAlreadyExists
}


func (r *ingredientRepository) Delete(ctx context.Context, id string) error {
	filter := map[string]any{"id": id}
	updates := map[string]any{
		"is_deleted": true,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		r.logger.Error("failed to delete ingredient", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *ingredientRepository) UnDelete(ctx context.Context, id string) error {
	filter := map[string]any{"id": id}
	updates := map[string]any{
		"is_deleted": false,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		r.logger.Error("failed to undelete ingredient", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}