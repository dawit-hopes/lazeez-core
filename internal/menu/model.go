package menu

import (
	"lazeez-core/internal/common"
	"mime/multipart"
)

type Menu struct {
	common.Base
	Name        string `json:"name" db:"name"`
	Image       string `json:"image" db:"image"`
	Description string `json:"description" db:"description"`
	
}

type MenuRequest struct {
	Name        string               `json:"name"`
	ImageHeader multipart.FileHeader `json:"-"`
	Image       multipart.File       `json:"image"`
}

func (m *Menu) Table() string {
	return "menus"
}

func (m *Menu) Columns() []string {
	return []string{"id", "name", "image", "deleted_at", "is_deleted"}
}

func (m *Menu) Values() []any {
	return []any{m.ID, m.Name, m.Image, m.DeletedAt, m.IsDeleted}
}

func (m *Menu) Addr() []any {
	return []any{&m.ID, &m.Name, &m.Image, &m.DeletedAt, &m.IsDeleted, &m.CreatedAt, &m.UpdatedAt}
}
