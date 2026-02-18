package ingredient

import "lazeez-core/internal/common"

type IngredientDTO struct {
	common.BaseDTO
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type IngredientRequest struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}
