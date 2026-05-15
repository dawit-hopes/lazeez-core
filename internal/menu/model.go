package menu

import (
	"database/sql"
	"lazeez-core/internal/common"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Menu struct {
	common.Base
	Name            string         `json:"name" db:"name"`
	Image           string         `json:"image" db:"image"`
	Description     string         `json:"description" db:"description"`
	Price           float64        `json:"price" db:"price"`
	Ingredients     pq.StringArray `json:"ingredients" db:"ingredients"`
	CategoryID      uuid.UUID      `json:"category_id" db:"category_id"`
	BranchID        sql.NullString `json:"branch_id" db:"branch_id"`
	MerchantID      sql.NullString `json:"merchant_id" db:"merchant_id"`
	IsFasting       bool           `json:"is_fasting" db:"is_fasting"`
	IsAvailable     bool           `json:"is_available" db:"is_available"`
	PreparationTime float64        `json:"preparation_time" db:"preparation_time"`
	Modifiers       pq.StringArray `json:"modifiers" db:"modifiers"`
	// Excluded is set only on branch list queries (not persisted on menus table).
	Excluded bool `json:"-" db:"-"`
}

func (m *Menu) Table() string {
	return "menus"
}

func (m *Menu) Columns() []string {
	return []string{
		"id", "name", "image", "deleted_at", "is_deleted", "branch_id", "merchant_id",
		"is_fasting", "is_available", "description", "price", "ingredients", "category_id", "modifiers", "preparation_time",
	}
}

func (m *Menu) Values() []any {
	return []any{
		m.ID, m.Name, m.Image, m.DeletedAt, m.IsDeleted, m.BranchID, m.MerchantID,
		m.IsFasting, m.IsAvailable, m.Description, m.Price, m.Ingredients, m.CategoryID, m.Modifiers, m.PreparationTime,
	}
}

func (m *Menu) Addr() []any {
	return []any{
		&m.ID, &m.Name, &m.Image, &m.DeletedAt, &m.IsDeleted, &m.BranchID, &m.MerchantID,
		&m.IsFasting, &m.IsAvailable, &m.Description, &m.Price, &m.Ingredients, &m.CategoryID, &m.Modifiers, &m.PreparationTime,
		&m.CreatedAt, &m.UpdatedAt,
	}
}

func (m *Menu) IsMaster() bool {
	return !m.BranchID.Valid || m.BranchID.String == ""
}

func (m *Menu) BranchIDString() string {
	if m.BranchID.Valid {
		return m.BranchID.String
	}
	return ""
}

func (m *Menu) MerchantIDString() string {
	if m.MerchantID.Valid {
		return m.MerchantID.String
	}
	return ""
}

func (m *Menu) ToDTO() MenuDTO {
	return MenuDTO{
		BaseDTO: common.BaseDTO{
			ID:        m.ID,
			IsDeleted: m.IsDeleted,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			DeletedAt: common.ToNullTimePtr(m.DeletedAt),
		},
		Name:            m.Name,
		Image:           m.Image,
		Description:     m.Description,
		Price:           m.Price,
		CategoryID:      common.ParseUUIDToString(m.CategoryID),
		BranchID:        m.BranchIDString(),
		MerchantID:      m.MerchantIDString(),
		IsMaster:        m.IsMaster(),
		IsExcluded:      m.Excluded,
		IsFasting:       m.IsFasting,
		IsAvailable:     m.IsAvailable,
		PreparationTime: m.PreparationTime,
	}
}
