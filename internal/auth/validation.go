package auth

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (u *LoginRequest) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.PhoneNumber,
			validation.Required.Error("phone number is required"),
			validation.Match(regexp.MustCompile(`^\+?251[79]\d{8}$`)).Error("phone number must be a valid Ethiopian phone number")),
		validation.Field(&u.Password, validation.Required.Error("password is required"), validation.Length(8, 100).Error("password must be between 8 and 100 characters")),
	)
}

func (u *SetPasswordRequest) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.PhoneNumber,
			validation.Required.Error("phone number is required"),
			validation.Match(regexp.MustCompile(`^\+?251[79]\d{8}$`)).Error("phone number must be a valid Ethiopian phone number")),
		validation.Field(&u.Password, validation.Required.Error("password is required"), validation.Length(8, 100).Error("password must be between 8 and 100 characters")),
	)
}
