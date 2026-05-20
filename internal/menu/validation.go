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
		validation.Field(&r.Description,
			validation.When(r.Description != "",
				validation.Length(3, 1000).Error("description must be between 3 and 1000 characters"),
			),
		),
		validation.Field(&r.Price,
			validation.Required.Error("price is required"),
			validation.Min(0.0).Error("price must be greater than 0"),
		),
		validation.Field(&r.Ingredients,
			validation.When(len(r.Ingredients) > 0,
				validation.Each(validation.Required.Error("ingredient is required")),
			),
		),
		validation.Field(&r.CategoryID,
			validation.Required.Error("category id is required"),
		),
	)
}
