package option

import "lazeez-core/internal/common"

type ModifierOptionDTO struct {
	common.BaseDTO
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
