package room

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (r *RoomRequestDTO) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.RoomNumber, validation.Required.Error("room number is required"), validation.Length(1, 100).Error("room number must be between 1 and 100 characters")),
		validation.Field(&r.RoomTypeID, validation.Required.Error("room type id is required")),
		validation.Field(&r.Floor, validation.Min(0).Error("floor must be 0 or greater")),
	)
}

func (r *RoomUpdateRequestDTO) Validate() error {
	if r.IsEmpty() {
		return validation.NewError("validation", "at least one field is required")
	}
	rules := []*validation.FieldRules{}
	if r.RoomNumber != "" {
		rules = append(rules, validation.Field(&r.RoomNumber, validation.Length(1, 100).Error("room number must be between 1 and 100 characters")))
	}
	if r.Floor != nil {
		rules = append(rules, validation.Field(r.Floor, validation.Min(0).Error("floor must be 0 or greater")))
	}
	return validation.ValidateStruct(r, rules...)
}

func (r *RoomUpdateRequestDTO) IsEmpty() bool {
	return r.RoomNumber == "" && r.RoomTypeID == "" && r.Floor == nil
}
