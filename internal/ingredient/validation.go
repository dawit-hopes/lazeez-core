package ingredient

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (r *IngredientRequest) Validate(requireName bool) error {
	rules := []*validation.FieldRules{
		validation.Field(&r.Name,
			validation.When(requireName, validation.Required.Error("name is required")),
			validation.When(r.Name != "", validation.Length(1, 100).Error("name must be between 1 and 100 characters")),
		),
	}
	return validation.ValidateStruct(r, rules...)
}
