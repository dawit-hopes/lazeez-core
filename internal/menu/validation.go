package menu

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (r *MenuRequest) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.Name,
			validation.Required.Error("name is required"),
			validation.Length(3, 100).Error("name must be between 3 and 100 characters"),
		),
		validation.Field(&r.ImageHeader,
			validation.Required.Error("image header is required"),
		),
		validation.Field(&r.Image,
			validation.Required.Error("image is required"),
		),
	)
}
