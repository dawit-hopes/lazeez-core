package merchant

import "lazeez-core/internal/common"

type Merchant struct {
	common.Base
	Name string `json:"name" db:"name"`
}

func (m *Merchant) Table() string {
	return "merchants"
}

func (m *Merchant) Columns() []string {
	return []string{"id", "name", "created_at", "updated_at", "deleted_at", "is_deleted"}
}

func (m *Merchant) Values() []any {
	return []any{m.ID, m.Name, m.CreatedAt, m.UpdatedAt, m.DeletedAt, m.IsDeleted}
}

func (m *Merchant) Addr() []any {
	return []any{&m.ID, &m.Name, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt, &m.IsDeleted}
}
