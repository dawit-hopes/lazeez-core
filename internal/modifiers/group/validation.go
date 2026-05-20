package group

import (
	"errors"

	"lazeez-core/internal/common"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func flexIntAtLeast(min int, message string) validation.RuleFunc {
	return func(value any) error {
		n, ok := value.(common.FlexInt)
		if !ok {
			return nil
		}
		if n.Int() < min {
			return errors.New(message)
		}
		return nil
	}
}

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
			validation.By(flexIntAtLeast(0, "min selections must be at least 0")),
		),
		validation.Field(&r.MaxSelections,
			validation.By(flexIntAtLeast(0, "max selections must be at least 0")),
		),
	)
}
