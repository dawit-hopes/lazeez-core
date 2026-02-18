package option

type ModifierOptionDTO struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	PriceAdjustment float64 `json:"price_adjustment"`
	IsDefault       bool    `json:"is_default"`
	IsAvailable     bool    `json:"is_available"`
}

type ModifierOptionRequest struct {
	Name            string  `json:"name"`
	PriceAdjustment float64 `json:"price_adjustment"`
	IsDefault       bool    `json:"is_default"`
	IsAvailable     bool    `json:"is_available"`
}
