package group

import (
	"lazeez-core/internal/common"
	"lazeez-core/internal/modifiers/option"

	"github.com/lib/pq"
)

type SelectionType string

const (
	SelectionTypeSingle   SelectionType = "single"
	SelectionTypeMultiple SelectionType = "multiple"
)

type ModifierGroup struct {
	common.Base
	Name          string         `json:"name" db:"name"`
	SelectionType string         `json:"selection_type" db:"selection_type"`
	IsRequired    bool           `json:"is_required" db:"is_required"`
	MinSelections int            `json:"min_selections" db:"min_selections"`
	MaxSelections int            `json:"max_selections" db:"max_selections"`
	Options       pq.StringArray `json:"options" db:"options"`
}

func (s *ModifierGroup) Table() string {
	return "modifier_groups"
}

func (s *ModifierGroup) Columns() []string {
	return []string{"id", "name", "selection_type", "is_required", "min_selections", "max_selections", "options", "deleted_at", "is_deleted"}
}

func (s *ModifierGroup) Values() []any {
	return []any{s.ID, s.Name, s.SelectionType, s.IsRequired, s.MinSelections, s.MaxSelections, s.Options, s.DeletedAt, s.IsDeleted}
}

func (s *ModifierGroup) Addr() []any {
	return []any{&s.ID, &s.Name, &s.SelectionType, &s.IsRequired, &s.MinSelections, &s.MaxSelections, &s.Options, &s.DeletedAt, &s.IsDeleted, &s.CreatedAt, &s.UpdatedAt}
}

func (s *ModifierGroup) ToDTO(options []option.ModifierOptionDTO) ModifierGroupDTO {
	return ModifierGroupDTO{
		ID:            s.ID,
		Name:          s.Name,
		SelectionType: SelectionType(s.SelectionType),
		IsRequired:    s.IsRequired,
		MinSelections: common.FlexInt(s.MinSelections),
		MaxSelections: common.FlexInt(s.MaxSelections),
		Options:       options,
	}
}
