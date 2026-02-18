package group

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ModifierGroup struct {
	ID            uuid.UUID      `db:"id"`
	Name          string         `db:"name"`
	SelectionType string         `db:"selection_type"`
	IsRequired    bool           `db:"is_required"`
	MinSelections int            `db:"min_selections"`
	MaxSelections int            `db:"max_selections"`
	Options       pq.StringArray `db:"options"`
	CreatedAt     time.Time      `db:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at"`
}

// ModifierOption represents a selectable option within a modifier group

type SelectionType string

var (
	SelectionTypeSingle   SelectionType = "single"
	SelectionTypeMultiple SelectionType = "multiple"
)

func (s *ModifierGroup) Table() string {
	return "modifier_groups"
}

func (s *ModifierGroup) Columns() []string {
	return []string{"id", "name", "selection_type", "is_required", "min_selections", "max_selections", "options"}
}

func (s *ModifierGroup) Values() []any {
	return []any{s.ID, s.Name, s.SelectionType, s.IsRequired, s.MinSelections, s.MaxSelections, s.Options}
}

func (s *ModifierGroup) Addr() []any {
	return []any{&s.ID, &s.Name, &s.SelectionType, &s.IsRequired, &s.MinSelections, &s.MaxSelections, &s.Options, &s.CreatedAt, &s.UpdatedAt}
}
