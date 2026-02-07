package merchant

import "lazeez-core/internal/common"

type Merchant struct {
	common.Base
	Name string `json:"name" db:"name"`
	Logo string `json:"logo" db:"logo"`
}

func (m *Merchant) Table() string {
	return "merchants"
}

func (m *Merchant) Columns() []string {
	return []string{"id", "name", "logo", "created_at", "updated_at", "deleted_at", "is_deleted"}
}

func (m *Merchant) Values() []any {
	return []any{m.ID, m.Name, m.Logo, m.CreatedAt, m.UpdatedAt, m.DeletedAt, m.IsDeleted}
}

func (m *Merchant) Addr() []any {
	return []any{&m.ID, &m.Name, &m.Logo, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt, &m.IsDeleted}
}
