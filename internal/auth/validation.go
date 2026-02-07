package auth

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (u *UserRequest) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.PhoneNumber,
			validation.Required.Error("phone number is required"),
			validation.Match(regexp.MustCompile(`^251[79]\d{8}$`)).Error("phone number must be a valid Ethiopian phone number"),
		),
		validation.Field(&u.FullName,
			validation.Required.Error("full name is required"),
			validation.Length(3, 100).Error("full name must be between 3 and 100 characters")),
		validation.Field(&u.BranchID, validation.Required.Error("branch id is required"),
			validation.Match(regexp.MustCompile(`^[0-9a-f-]+$`)).Error("branch id must be a valid UUID")),
		validation.Field(&u.MerchantID, validation.Required.Error("merchant id is required"),
			validation.Match(regexp.MustCompile(`^[0-9a-f-]+$`)).Error("merchant id must be a valid UUID")),
	)
}

func (u *SetPasswordRequest) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.PhoneNumber, validation.Required.Error("phone number is required"), validation.Match(regexp.MustCompile(`^[0-9]+$`)).Error("phone number must be a number")),
		validation.Field(&u.Password, validation.Required.Error("password is required"), validation.Length(8, 100).Error("password must be between 8 and 100 characters")),
	)
}

func (u *LoginRequest) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.PhoneNumber, validation.Required.Error("phone number is required"), validation.Match(regexp.MustCompile(`^[0-9]+$`)).Error("phone number must be a number")),
		validation.Field(&u.Password, validation.Required.Error("password is required"), validation.Length(8, 100).Error("password must be between 8 and 100 characters")),
	)
}

func (u *UserLookUpRequest) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.PhoneNumber, validation.Required.Error("phone number is required"),
			validation.Match(regexp.MustCompile(`^[0-9]+$`)).Error("phone number must be a number"),
			validation.Match(regexp.MustCompile(`^251[79]\d{8}$`)).Error("phone number must be a valid Ethiopian phone number")),
	)
}
