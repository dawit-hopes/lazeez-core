package rooms

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (r *RoomRequestDTO) Validate(isUpdate bool) error {
	rules := []*validation.FieldRules{
		validation.Field(&r.Name, validation.Required.Error("name is required"), validation.Length(3, 100).Error("name must be between 3 and 100 characters")),
		validation.Field(&r.Description, validation.Required.Error("description is required"), validation.Length(3, 1000).Error("description must be between 3 and 1000 characters")),
		validation.Field(&r.PricePerNight, validation.Required.Error("price per night is required"), validation.Min(0.01).Error("price per night must be greater than 0")),
	}
	if !isUpdate {
		rules = append(rules, validation.Field(&r.MerchantID, validation.Required.Error("merchant id is required")))
	}
	return validation.ValidateStruct(r, rules...)
}

func (r *RoomRequestDTO) ValidateUpdate() error {
	if r.IsEmpty() {
		return validation.NewError("validation", "at least one field is required")
	}
	rules := []*validation.FieldRules{}
	if r.Name != "" {
		rules = append(rules, validation.Field(&r.Name, validation.Length(3, 100).Error("name must be between 3 and 100 characters")))
	}
	if r.Description != "" {
		rules = append(rules, validation.Field(&r.Description, validation.Length(3, 1000).Error("description must be between 3 and 1000 characters")))
	}
	if r.PricePerNight != 0 {
		rules = append(rules, validation.Field(&r.PricePerNight, validation.Min(0.01).Error("price per night must be greater than 0")))
	}
	return validation.ValidateStruct(r, rules...)
}

func (r *RoomRequestDTO) IsEmpty() bool {
	return r.Name == "" && r.Description == "" && r.PricePerNight == 0
}

func (r *CloneRoomRequestDTO) Validate() error {
	rules := []*validation.FieldRules{}
	if r.Name != "" {
		rules = append(rules, validation.Field(&r.Name, validation.Length(3, 100).Error("name must be between 3 and 100 characters")))
	}
	if r.Description != "" {
		rules = append(rules, validation.Field(&r.Description, validation.Length(3, 1000).Error("description must be between 3 and 1000 characters")))
	}
	if r.PricePerNight != 0 {
		rules = append(rules, validation.Field(&r.PricePerNight, validation.Min(0.01).Error("price per night must be greater than 0")))
	}
	if len(rules) == 0 {
		return nil
	}
	return validation.ValidateStruct(r, rules...)
}
