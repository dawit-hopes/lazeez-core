package menu

import (
	"context"
	"database/sql"
	"errors"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type MenuRepository interface {
	Create(ctx context.Context, menu Menu) error
	Get(ctx context.Context, id string, branchID string) (Menu, error)
	Update(ctx context.Context, menu Menu) error
	Delete(ctx context.Context, id string, branchID string) error
	UnDelete(ctx context.Context, id string, branchID string) error
	List(ctx context.Context, filter common.Filter, branchID string) ([]*Menu, error)
	CheckExists(ctx context.Context, name, branchID string) error
}

type menuRepository struct {
	dal    *common.DAL[*Menu]
	logger config.Logger
}

func NewMenuRepository(dal *common.DAL[*Menu], logger config.Logger) MenuRepository {
	return &menuRepository{dal: dal, logger: logger}
}

func (r *menuRepository) Create(ctx context.Context, menu Menu) error {
	_, err := r.dal.Create(ctx, &menu)
	if err != nil {
		r.logger.Error("failed to create menu", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *menuRepository) Get(ctx context.Context, id string, branchID string) (Menu, error) {
	filter := map[string]any{"id": id, "branch_id": branchID}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("menu not found", "error", err)
			return Menu{}, common.ErrMenuNotFound
		}
		r.logger.Error("failed to get menu", "error", err)
		return Menu{}, common.ErrInternalServerError
	}
	return *result, nil
}

func (r *menuRepository) Update(ctx context.Context, menu Menu) error {
	filter := map[string]any{"id": menu.ID}
	updates := map[string]any{
		"name":         menu.Name,
		"image":        menu.Image,
		"description":  menu.Description,
		"price":        menu.Price,
		"ingredients":  menu.Ingredients,
		"category_id":  menu.CategoryID,
		"branch_id":    menu.BranchID,
		"is_fasting":   menu.IsFasting,
		"is_available": menu.IsAvailable,
		"deleted_at":   menu.DeletedAt,
		"is_deleted":   menu.IsDeleted,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("menu not found", "error", err)
			return common.ErrMenuNotFound
		}
		r.logger.Error("failed to update menu", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *menuRepository) Delete(ctx context.Context, id string, branchID string) error {
	filter := map[string]any{"id": id, "branch_id": branchID}
	updates := map[string]any{
		"is_deleted": true,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("menu not found", "error", err)
			return common.ErrMenuNotFound
		}
		r.logger.Error("failed to delete menu", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *menuRepository) UnDelete(ctx context.Context, id string, branchID string) error {
	filter := map[string]any{"id": id, "branch_id": branchID}
	updates := map[string]any{
		"is_deleted": false,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Error("menu not found", "error", err)
			return common.ErrMenuNotFound
		}
		r.logger.Error("failed to undelete menu", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *menuRepository) List(ctx context.Context, filter common.Filter, branchID string) ([]*Menu, error) {
	filters := map[string]any{"branch_id": branchID}
	if filter.Search != "" {
		filters["name"] = common.ILike(filter.Search)
	}
	menus, err := r.dal.List(ctx, filters, filter.Page, filter.Limit)
	if err != nil {
		r.logger.Error("failed to list menus", "error", err)
		return nil, common.ErrInternalServerError
	}
	return menus, nil
}

func (r *menuRepository) CheckExists(ctx context.Context, name, branchID string) error {
	filter := map[string]any{"name": name, "branch_id": branchID}
	_, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		r.logger.Error("failed to check if menu exists", "error", err)
		return common.ErrInternalServerError
	}
	return common.ErrMenuAlreadyExists
}
