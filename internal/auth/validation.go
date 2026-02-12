package auth

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (u *LoginRequest) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.PhoneNumber,
			validation.Required.Error("phone number is required"),
		),
	)
}

func (u *SetPasswordRequest) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.PhoneNumber,
			validation.Required.Error("phone number is required"),
		),
	)
}
