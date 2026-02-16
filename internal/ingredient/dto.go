package ingredient

import "lazeez-core/internal/common"

type IngredientDTO struct {
	common.BaseDTO
	Name string `json:"name"`
}
