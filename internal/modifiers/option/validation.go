package option

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (r *ModifierOptionRequest) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.Name,
			validation.Required.Error("name is required"),
			validation.Length(3, 100).Error("name must be between 3 and 100 characters"),
		),
		validation.Field(&r.PriceAdjustment,
			validation.Min(0).Error("price adjustment must be greater than 0"),
		),
		// IsDefault and IsAvailable are bools; false is valid, so we don't use Required here.
	)
}
