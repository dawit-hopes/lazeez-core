package group

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (r *ModifierGroupRequest) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.Name,
			validation.Required.Error("name is required"),
			validation.Length(3, 100).Error("name must be between 3 and 100 characters"),
		),
		validation.Field(&r.SelectionType,
			validation.Required.Error("selection type is required"),
			validation.In(SelectionTypeSingle, SelectionTypeMultiple).Error("selection type must be single or multiple"),
		),
		validation.Field(&r.MinSelections,
			validation.Required.Error("min selections is required"),
			validation.Min(0).Error("min selections must be greater than 0"),
		),
		validation.Field(&r.MaxSelections,
			validation.Required.Error("max selections is required"),
			validation.Min(0).Error("max selections must be greater than 0"),
		),
	)
}
