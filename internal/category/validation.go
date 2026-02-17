package category

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (r *CategoryRequest) Validate(requireIcon bool) error {
	rules := []*validation.FieldRules{
		validation.Field(&r.Name,
			validation.When(requireIcon, validation.Required.Error("name is required")),
			validation.When(r.Name != "", validation.Length(1, 100).Error("name must be between 1 and 100 characters")),
		),
	}
	if requireIcon {
		rules = append(rules,
			validation.Field(&r.IconHeader,
				validation.Required.Error("icon is required"),
			),
			validation.Field(&r.Icon,
				validation.Required.Error("icon file is required"),
			),
		)
	}
	return validation.ValidateStruct(r, rules...)
}
