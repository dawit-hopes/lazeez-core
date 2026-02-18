package option

import "lazeez-core/internal/common"

type ModifierOption struct {
	common.Base
	Name            string  `json:"name" db:"name"`
	PriceAdjustment float64 `json:"price_adjustment" db:"price_adjustment"`
	IsDefault       bool    `json:"is_default" db:"is_default"`
	IsAvailable     bool    `json:"is_available" db:"is_available"`
}

func (s *ModifierOption) Table() string {
	return "modifier_options"
}

func (s *ModifierOption) Columns() []string {
	return []string{"id", "name", "price_adjustment", "is_default", "is_available", "deleted_at", "is_deleted"}
}

func (s *ModifierOption) Values() []any {
	return []any{s.ID, s.Name, s.PriceAdjustment, s.IsDefault, s.IsAvailable, s.DeletedAt, s.IsDeleted}
}

func (s *ModifierOption) Addr() []any {
	return []any{&s.ID, &s.Name, &s.PriceAdjustment, &s.IsDefault, &s.IsAvailable, &s.DeletedAt, &s.IsDeleted, &s.CreatedAt, &s.UpdatedAt}
}

func (s *ModifierOption) ToDTO() ModifierOptionDTO {
	return ModifierOptionDTO{
		BaseDTO: common.BaseDTO{
			ID:        s.ID,
			IsDeleted: s.IsDeleted,
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
			DeletedAt: common.ToNullTimePtr(s.DeletedAt),
		},
		Name:            s.Name,
		PriceAdjustment: s.PriceAdjustment,
		IsDefault:       s.IsDefault,
		IsAvailable:     s.IsAvailable,
	}
}
