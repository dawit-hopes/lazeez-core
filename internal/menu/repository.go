package menu

import (
	"context"
	"database/sql"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type MenuRepository interface {
	Create(ctx context.Context, menu Menu) (Menu, error)
	Get(ctx context.Context, id string) (Menu, error)
	Update(ctx context.Context, menu Menu) (Menu, error)
	Delete(ctx context.Context, id string) error
}

type menuRepository struct {
	dal    *common.DAL[*Menu]
	logger config.Logger
}

func NewMenuRepository(dal *common.DAL[*Menu], logger config.Logger) MenuRepository {
	return &menuRepository{dal: dal, logger: logger}
}

func (r *menuRepository) Create(ctx context.Context, menu Menu) (Menu, error) {
	result, err := r.dal.Create(ctx, &menu)
	if err != nil {
		r.logger.Error("failed to create menu", "error", err)
		return menu, common.ErrInternalServerError
	}
	return *result, nil
}

func (r *menuRepository) Get(ctx context.Context, id string) (Menu, error) {
	filter := map[string]any{"id": id}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Error("menu not found", "error", err)
			return Menu{}, common.ErrMenuNotFound
		}
		r.logger.Error("failed to get menu", "error", err)
		return Menu{}, common.ErrInternalServerError
	}
	return *result, nil
}

func (r *menuRepository) Update(ctx context.Context, menu Menu) (Menu, error) {
	filter := map[string]any{"id": menu.ID}
	result, err := r.dal.Update(ctx, filter, &menu)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Error("menu not found", "error", err)
			return Menu{}, common.ErrMenuNotFound
		}
		r.logger.Error("failed to update menu", "error", err)
		return Menu{}, common.ErrInternalServerError
	}
	return *result, nil
}

func (r *menuRepository) Delete(ctx context.Context, id string) error {
	err := r.dal.Delete(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Error("menu not found", "error", err)
			return common.ErrMenuNotFound
		}
		r.logger.Error("failed to delete menu", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}
