package menu

import (
	"database/sql"
	"encoding/json"
)

type DiscountType string

const (
	DiscountTypePercentage DiscountType = "percentage"
	DiscountTypeFixed      DiscountType = "fixed"
)

type MenuDiscount struct {
	Type  DiscountType `json:"type"`
	Value float64      `json:"value"`
}

func (d *MenuDiscount) IsValid(price float64) bool {
	if d == nil {
		return true
	}
	switch d.Type {
	case DiscountTypePercentage:
		return d.Value > 0 && d.Value <= 100
	case DiscountTypeFixed:
		return d.Value > 0 && (price <= 0 || d.Value < price)
	default:
		return false
	}
}

func parseMenuDiscount(raw string) (*MenuDiscount, error) {
	if raw == "" || raw == "null" {
		return nil, nil
	}
	var discount MenuDiscount
	if err := json.Unmarshal([]byte(raw), &discount); err != nil {
		return nil, err
	}
	return &discount, nil
}

func discountFromModel(discountType sql.NullString, discountValue sql.NullFloat64) *MenuDiscount {
	if !discountType.Valid || !discountValue.Valid {
		return nil
	}
	return &MenuDiscount{
		Type:  DiscountType(discountType.String),
		Value: discountValue.Float64,
	}
}

func applyDiscountToModel(menu *Menu, discount *MenuDiscount) {
	if discount == nil {
		menu.DiscountType = sql.NullString{}
		menu.DiscountValue = sql.NullFloat64{}
		return
	}
	menu.DiscountType = sql.NullString{String: string(discount.Type), Valid: true}
	menu.DiscountValue = sql.NullFloat64{Float64: discount.Value, Valid: true}
}
