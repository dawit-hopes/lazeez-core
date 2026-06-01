package booking

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (b *BookingRequestDTO) Validate() error {
	return validation.ValidateStruct(b,
		validation.Field(&b.RoomID, validation.Required.Error("room id is required")),
		validation.Field(&b.GuestName, validation.Required.Error("guest name is required"), validation.Length(1, 255).Error("guest name must be between 1 and 255 characters")),
		validation.Field(&b.GuestPhone, validation.Required.Error("guest phone is required"), validation.Length(1, 50).Error("guest phone must be between 1 and 50 characters")),
		validation.Field(&b.NumberOfNights, validation.Required.Error("number of nights is required"), validation.Min(1).Error("number of nights must be at least 1")),
		validation.Field(&b.CheckInDate, validation.Required.Error("check in date is required")),
	)
}

func (b *BookingUpdateRequestDTO) Validate() error {
	if b.IsEmpty() {
		return validation.NewError("validation", "at least one field is required")
	}
	rules := []*validation.FieldRules{}
	if b.GuestName != "" {
		rules = append(rules, validation.Field(&b.GuestName, validation.Length(1, 255).Error("guest name must be between 1 and 255 characters")))
	}
	if b.GuestPhone != "" {
		rules = append(rules, validation.Field(&b.GuestPhone, validation.Length(1, 50).Error("guest phone must be between 1 and 50 characters")))
	}
	return validation.ValidateStruct(b, rules...)
}
