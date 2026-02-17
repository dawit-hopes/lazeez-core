package menu

import (
	"lazeez-core/internal/common"
)

type Menu struct {
	common.Base
	Name        string   `json:"name" db:"name"`
	Image       string   `json:"image" db:"image"`
	Description string   `json:"description" db:"description"`
	Price       float64  `json:"price" db:"price"`
	Ingredients []string `json:"ingredients" db:"ingredients"`
	CategoryID  string   `json:"category_id" db:"category_id"`
	BranchID    string   `json:"branch_id" db:"branch_id"`
	IsFasting   bool     `json:"is_fasting" db:"is_fasting"`
	IsAvailable bool     `json:"is_available" db:"is_available"`
}

func (m *Menu) Table() string {
	return "menus"
}

func (m *Menu) Columns() []string {
	return []string{"id", "name", "image", "deleted_at", "is_deleted", "branch_id", "is_fasting", "is_available", "description", "price", "ingredients", "category_id"}
}

func (m *Menu) Values() []any {
	return []any{m.ID, m.Name, m.Image, m.DeletedAt, m.IsDeleted, m.BranchID, m.IsFasting, m.IsAvailable, m.Description, m.Price, m.Ingredients, m.CategoryID}
}

func (m *Menu) Addr() []any {
	return []any{&m.ID, &m.Name, &m.Image, &m.DeletedAt, &m.IsDeleted, &m.BranchID, &m.IsFasting, &m.IsAvailable, &m.Description, &m.Price, &m.Ingredients, &m.CategoryID, &m.CreatedAt, &m.UpdatedAt}
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
		Name:        m.Name,
		Image:       m.Image,
		Description: m.Description,
		Price:       m.Price,
		Ingredients: m.Ingredients,
		CategoryID:  m.CategoryID,
		BranchID:    m.BranchID,
		IsFasting:   m.IsFasting,
		IsAvailable: m.IsAvailable,
	}
}
