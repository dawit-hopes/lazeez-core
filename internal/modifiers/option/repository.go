package option

import (
	"context"
	"database/sql"
	"errors"
	"lazeez-core/config"
	"lazeez-core/internal/common"

	"github.com/lib/pq"
)

type ModifierOptionRepository interface {
	Create(ctx context.Context, modifierOption ModifierOption) error
	Get(ctx context.Context, id string) (ModifierOption, error)
	GetByIDs(ctx context.Context, ids []string) (map[string]ModifierOption, error)
	Update(ctx context.Context, modifierOption ModifierOption) error
	Delete(ctx context.Context, id string) error
	UnDelete(ctx context.Context, id string) error
}

type modifierOptionRepository struct {
	dal    *common.DAL[*ModifierOption]
	join   *common.JoinDAL
	logger config.Logger
}

func NewModifierOptionRepository(dal *common.DAL[*ModifierOption], join *common.JoinDAL, logger config.Logger) ModifierOptionRepository {
	return &modifierOptionRepository{dal: dal, join: join, logger: logger}
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
	filter := map[string]any{"id": id, "is_deleted": false}
	result, err := r.dal.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ModifierOption{}, common.ErrModifierOptionNotFound
		}
		r.logger.Error("failed to get modifier option", "error", err)
		return ModifierOption{}, common.ErrInternalServerError
	}
	return *result, nil
}

const getModifierOptionsByIDsQuery = `
SELECT id, name, price_adjustment, is_default, is_available, deleted_at, is_deleted, created_at, updated_at
FROM modifier_options
WHERE is_deleted = FALSE AND id = ANY($1)`

func (r *modifierOptionRepository) GetByIDs(ctx context.Context, ids []string) (map[string]ModifierOption, error) {
	if len(ids) == 0 {
		r.logger.Debug("modifier option batch lookup skipped", "reason", "empty id list")
		return map[string]ModifierOption{}, nil
	}

	r.logger.Debug("loading modifier options for batch lookup", "requested_count", len(ids))

	rows, err := common.QueryRows(r.join, ctx, getModifierOptionsByIDsQuery, []any{pq.Array(ids)}, func(rows *sql.Rows) (ModifierOption, error) {
		var option ModifierOption
		if err := rows.Scan(option.Addr()...); err != nil {
			return ModifierOption{}, err
		}
		return option, nil
	})
	if err != nil {
		r.logger.Error(
			"failed to get modifier options by ids",
			"requested_count", len(ids),
			"error", err,
		)
		return nil, common.ErrInternalServerError
	}

	result := make(map[string]ModifierOption, len(rows))
	for _, option := range rows {
		result[option.ID] = option
	}

	if len(result) < len(ids) {
		r.logger.Debug(
			"modifier option batch lookup returned partial results",
			"requested_count", len(ids),
			"found_count", len(result),
		)
	}

	r.logger.Debug(
		"loaded modifier options for batch lookup",
		"requested_count", len(ids),
		"found_count", len(result),
	)
	return result, nil
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
