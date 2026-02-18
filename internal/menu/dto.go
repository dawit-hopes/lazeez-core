package menu

import (
	"lazeez-core/internal/category"
	"lazeez-core/internal/common"
	"lazeez-core/internal/ingredient"
	"lazeez-core/internal/modifiers/group"
	"mime/multipart"

	"github.com/lib/pq"
)

type MenuRequest struct {
	Name           string                       `json:"name"`
	ImageHeader    multipart.FileHeader         `json:"-"`
	Image          multipart.File               `json:"image"`
	Description    string                       `json:"description"`
	Price          float64                      `json:"price"`
	Ingredients    []string                     `json:"ingredients"`
	CategoryID     string                       `json:"category_id"`
	BranchID       string                       `json:"branch_id"`
	IsFasting      *bool                        `json:"is_fasting"`
	IsAvailable    *bool                        `json:"is_available"`
	Modifiers      []group.ModifierGroupRequest `json:"modifier_groups"`
}

type MenuDTO struct {
	common.BaseDTO
	Name        string                      `json:"name"`
	Image       string                      `json:"image"`
	Description string                      `json:"description"`
	Price       float64                     `json:"price"`
	CategoryID  string                      `json:"category_id"`
	Category    *category.CategoryDTO       `json:"category"`
	Ingredients []*ingredient.IngredientDTO `json:"ingredients"`
	BranchID    string                      `json:"branch_id"`
	IsFasting   bool                        `json:"is_fasting"`
	IsAvailable bool                        `json:"is_available"`
	Modifiers   []group.ModifierGroupDTO    `json:"modifier_groups"`
}

func (m *MenuRequest) IsEmpty() bool {
	return m.Name == "" && m.Image == nil && m.Description == "" && m.Price == 0 && len(m.Ingredients) == 0 && m.CategoryID == "" && m.BranchID == "" && m.IsFasting == nil && m.IsAvailable == nil
}

func (m *MenuRequest) ToModel() Menu {
	ingredients := make(pq.StringArray, len(m.Ingredients))
	copy(ingredients, m.Ingredients)
	return Menu{
		Name:        common.FormatText(m.Name),
		Description: m.Description,
		Price:       m.Price,
		Ingredients: ingredients,
		CategoryID:  common.ParseStringToUUID(m.CategoryID),
		BranchID:    common.ParseStringToUUID(m.BranchID),
		IsFasting:   *m.IsFasting,
		IsAvailable: *m.IsAvailable,
	}
}
