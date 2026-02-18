package ingredient

import "lazeez-core/internal/common"

type Ingredient struct {
	common.Base
	Name string `json:"name" db:"name"`
	Icon string `json:"icon" db:"icon"`
}

func (i *Ingredient) Table() string {
	return "ingredients"
}

func (i *Ingredient) Columns() []string {
	return []string{"id", "name", "icon", "deleted_at", "is_deleted"}
}

func (i *Ingredient) Values() []any {
	return []any{i.ID, i.Name, i.Icon, i.DeletedAt, i.IsDeleted}
}

func (i *Ingredient) Addr() []any {
	return []any{&i.ID, &i.Name, &i.Icon, &i.DeletedAt, &i.IsDeleted, &i.CreatedAt, &i.UpdatedAt}
}

func (i *Ingredient) ToDTO() IngredientDTO {
	return IngredientDTO{
		BaseDTO: common.BaseDTO{
			ID:        i.ID,
			IsDeleted: i.IsDeleted,
			CreatedAt: i.CreatedAt,
			UpdatedAt: i.UpdatedAt,
		},
		Name: i.Name,
		Icon: i.Icon,
	}
}
