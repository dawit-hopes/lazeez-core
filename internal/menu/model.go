package menu

import "lazeez-core/internal/common"

type Menu struct {
	common.Base
	Name string `json:"name" db:"name"`
}

func (m *Menu) Table() string {
	return "menus"
}

func (m *Menu) Columns() []string {
	return []string{"id", "name", "created_at", "updated_at", "deleted_at", "is_deleted"}
}

func (m *Menu) Values() []any {
	return []any{m.ID, m.Name, m.CreatedAt, m.UpdatedAt, m.DeletedAt, m.IsDeleted}
}

func (m *Menu) Addr() []any {
	return []any{&m.ID, &m.Name, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt, &m.IsDeleted}
}
