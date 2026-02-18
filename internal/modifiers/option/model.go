package option

import (
	"lazeez-core/internal/common"
	"time"

	"github.com/google/uuid"
)

type ModifierOption struct {
	ID              uuid.UUID `db:"id"`
	Name            string    `db:"name"`
	PriceAdjustment float64   `db:"price_adjustment"`
	IsDefault       bool      `db:"is_default"`
	IsAvailable     bool      `db:"is_available"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

func (s *ModifierOption) Table() string {
	return "modifier_options"
}

func (s *ModifierOption) Columns() []string {
	return []string{"id", "name", "price_adjustment", "is_default", "is_available"}
}

func (s *ModifierOption) Values() []any {
	return []any{s.ID, s.Name, s.PriceAdjustment, s.IsDefault, s.IsAvailable}
}

func (s *ModifierOption) Addr() []any {
	return []any{&s.ID, &s.Name, &s.PriceAdjustment, &s.IsDefault, &s.IsAvailable, &s.CreatedAt, &s.UpdatedAt}
}

func (s *ModifierOption) ToDTO() ModifierOptionDTO {
	return ModifierOptionDTO{
		ID:              common.ParseUUIDToString(s.ID),
		Name:            s.Name,
		PriceAdjustment: s.PriceAdjustment,
		IsDefault:       s.IsDefault,
		IsAvailable:     s.IsAvailable,
	}
}
