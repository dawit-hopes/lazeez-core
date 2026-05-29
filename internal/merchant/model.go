package merchant

import "lazeez-core/internal/common"

type Merchant struct {
	common.Base
	Name       string     `json:"name" db:"name"`
	BranchType BranchType `json:"branch_type" db:"branch_type"`
	Logo string `json:"logo" db:"logo"`
}

func (m *Merchant) Table() string {
	return "merchants"
}

func (m *Merchant) Columns() []string {
	return []string{"id", "name", "branch_type", "logo", "deleted_at", "is_deleted"}
}

func (m *Merchant) Values() []any {
	return []any{m.ID, m.Name, m.BranchType, m.Logo, m.DeletedAt, m.IsDeleted}
}

func (m *Merchant) Addr() []any {
	return []any{&m.ID, &m.Name, &m.BranchType, &m.Logo, &m.DeletedAt, &m.IsDeleted, &m.CreatedAt, &m.UpdatedAt}
}

func (m *Merchant) ToDTO() MerchantDTO {
	return MerchantDTO{
		BaseDTO: common.BaseDTO{
			ID:        m.ID,
			IsDeleted: m.IsDeleted,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			DeletedAt: common.ToNullTimePtr(m.DeletedAt),
		},
		Name: m.Name,
		BranchType: m.BranchType,
		Logo: m.Logo,
	}
}
