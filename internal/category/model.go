package category

import (
	"lazeez-core/internal/common"
)

type Category struct {
	common.Base
	Name string `json:"name" db:"name"`
	Icon string `json:"icon" db:"icon"`
	// Station routes this category's items to a preparation station: "kitchen" or "bar". May be empty.
	Station string `json:"station" db:"station"`
}

func (c *Category) Table() string {
	return "categories"
}

func (c *Category) Columns() []string {
	return []string{"id", "name", "icon", "station", "deleted_at", "is_deleted"}
}

func (c *Category) Values() []any {
	return []any{c.ID, c.Name, c.Icon, c.Station, c.DeletedAt, c.IsDeleted}
}

func (c *Category) Addr() []any {
	return []any{&c.ID, &c.Name, &c.Icon, &c.Station, &c.DeletedAt, &c.IsDeleted, &c.CreatedAt, &c.UpdatedAt}
}

func (c *Category) ToDTO() CategoryDTO {
	return CategoryDTO{
		BaseDTO: common.BaseDTO{
			ID:        c.ID,
			IsDeleted: c.IsDeleted,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			DeletedAt: common.ToNullTimePtr(c.DeletedAt),
		},
		Name:    c.Name,
		Icon:    c.Icon,
		Station: c.Station,
	}
}


func (c *Category) ToResponseSimplified() *CategoryResponseSimplified {
	return &CategoryResponseSimplified{
		Name:    c.Name,
		Icon:    c.Icon,
		Station: c.Station,
	}
}