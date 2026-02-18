package group

import (
	"context"
	"database/sql"
	"errors"
	"lazeez-core/config"
	"lazeez-core/internal/common"
)

type ModifierGroupRepository interface {
	Create(ctx context.Context, modifier ModifierGroup) error
	Get(ctx context.Context, id string) (ModifierGroup, error)
	Update(ctx context.Context, modifier ModifierGroup) error
	Delete(ctx context.Context, id string) error
	UnDelete(ctx context.Context, id string) error
}

type modifierRepository struct {
	dal    *common.DAL[*ModifierGroup]
	logger config.Logger
}

func NewModifierGroupRepository(dal *common.DAL[*ModifierGroup], logger config.Logger) ModifierGroupRepository {
	return &modifierRepository{dal: dal, logger: logger}
}

func (r *modifierRepository) Create(ctx context.Context, modifier ModifierGroup) error {
	_, err := r.dal.Create(ctx, &modifier)
	if err != nil {
		r.logger.Error("failed to create modifier", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *modifierRepository) Get(ctx context.Context, id string) (ModifierGroup, error) {
	filter := map[string]any{"id": id, "is_deleted": false}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ModifierGroup{}, common.ErrModifierGroupNotFound
		}
		r.logger.Error("failed to get modifier group", "error", err)
		return ModifierGroup{}, common.ErrInternalServerError
	}
	return *result, nil
}

func (r *modifierRepository) Update(ctx context.Context, modifier ModifierGroup) error {
	filter := map[string]any{"id": modifier.ID}
	updates := map[string]any{
		"name":           modifier.Name,
		"selection_type": modifier.SelectionType,
		"is_required":    modifier.IsRequired,
		"min_selections": modifier.MinSelections,
		"max_selections": modifier.MaxSelections,
		"options":        modifier.Options,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		r.logger.Error("failed to update modifier", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *modifierRepository) Delete(ctx context.Context, id string) error {
	filter := map[string]any{"id": id}
	updates := map[string]any{
		"is_deleted": true,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		r.logger.Error("failed to delete modifier", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}

func (r *modifierRepository) UnDelete(ctx context.Context, id string) error {
	filter := map[string]any{"id": id}
	updates := map[string]any{
		"is_deleted": false,
	}
	err := r.dal.Update(ctx, filter, updates)
	if err != nil {
		r.logger.Error("failed to undelete modifier", "error", err)
		return common.ErrInternalServerError
	}
	return nil
}
