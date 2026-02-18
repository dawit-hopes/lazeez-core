package option

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type ModifierOptionRepository interface {
	Create(ctx context.Context, modifierOption ModifierOption) error
	Get(ctx context.Context, id string) (ModifierOption, error)
	Update(ctx context.Context, modifierOption ModifierOption) error
	Delete(ctx context.Context, id string) error
	UnDelete(ctx context.Context, id string) error
}

type modifierOptionRepository struct {
	dal    *common.DAL[*ModifierOption]
	logger config.Logger
}

func NewModifierOptionRepository(dal *common.DAL[*ModifierOption], logger config.Logger) ModifierOptionRepository {
	return &modifierOptionRepository{dal: dal, logger: logger}
}

func (r *modifierOptionRepository) Create(ctx context.Context, modifierOption ModifierOption) error {
	_, err := r.dal.Create(ctx, &modifierOption)
	if err != nil {
		r.logger.Error("failed to create modifier option", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *modifierOptionRepository) Get(ctx context.Context, id string) (ModifierOption, error) {
	filter := map[string]any{"id": id}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		r.logger.Error("failed to get modifier option", "error", err)
		return ModifierOption{}, common.ErrInternalServerError
	}
	return *result, nil
}

func (r *modifierOptionRepository) Update(ctx context.Context, modifierOption ModifierOption) error {
	filter := map[string]any{"id": modifierOption.ID}
	updates := map[string]any{
		"name":             modifierOption.Name,
		"price_adjustment": modifierOption.PriceAdjustment,
		"is_default":       modifierOption.IsDefault,
		"is_available":     modifierOption.IsAvailable,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		r.logger.Error("failed to update modifier option", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *modifierOptionRepository) Delete(ctx context.Context, id string) error {
	filter := map[string]any{"id": id}
	updates := map[string]any{
		"is_deleted": true,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		r.logger.Error("failed to delete modifier option", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *modifierOptionRepository) UnDelete(ctx context.Context, id string) error {
	filter := map[string]any{"id": id}
	updates := map[string]any{
		"is_deleted": false,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		r.logger.Error("failed to undelete modifier option", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}
