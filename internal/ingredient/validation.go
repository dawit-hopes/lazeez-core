package ingredient

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (r *IngredientRequest) Validate(requireName, requireIcon bool) error {
	rules := []*validation.FieldRules{
		validation.Field(&r.Name,
			validation.When(requireName, validation.Required.Error("name is required")),
			validation.When(r.Name != "", validation.Length(1, 100).Error("name must be between 1 and 100 characters")),
		),
		validation.Field(&r.Icon,
			validation.When(requireIcon, validation.Required.Error("icon is required")),
			validation.When(r.Icon != "", validation.Length(1, 100).Error("icon must be between 1 and 100 characters")),
		),
	}
	return validation.ValidateStruct(r, rules...)
}

func (r *IngredientRequest) IsEmpty() bool {
	return r.Name == "" && r.Icon == ""
}
