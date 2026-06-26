package users

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Validate checks UserRequest. Phone format is validated and normalized in the handler once; here we only check required/length/etc.
func (u *UserRequest) Validate(isUpdate bool) error {
	return validation.ValidateStruct(u,
		validation.Field(&u.PhoneNumber,
			validation.When(!isUpdate, validation.Required.Error("phone number is required")),
		),
		validation.Field(&u.FullName,
			validation.When(!isUpdate, validation.Required.Error("full name is required")),
			validation.Length(3, 100).Error("full name must be between 3 and 100 characters")),
		validation.Field(&u.BranchID,
			validation.When(!isUpdate, validation.Required.Error("branch id is required")),
			validation.Match(regexp.MustCompile(`^[0-9a-f-]+$`)).Error("branch id must be a valid UUID")),
		validation.Field(&u.MerchantID,
			validation.When(!isUpdate, validation.Required.Error("merchant id is required")),
			validation.Match(regexp.MustCompile(`^[0-9a-f-]+$`)).Error("merchant id must be a valid UUID")),
		validation.Field(&u.Role,
			validation.When(!isUpdate, validation.Required.Error("role is required")),
			validation.In(
				RoleBranchManager, RoleFrontDeskAgent, RoleRoomServiceStaff,
				RoleWaiter, RoleKitchenStaff, RoleBarista, RoleCashier,
			).Error("role must be a valid role")),
		validation.Field(&u.Pin,
			validation.When(u.Role == RoleWaiter,
				validation.When(!isUpdate, validation.Required.Error("pin is required for waiters")),
				validation.Match(regexp.MustCompile(`^[0-9]{4,6}$`)).Error("pin must be 4 to 6 digits")),
		),
	)
}

func (u *SuperAdminUserRequest) Validate(isUpdate bool) error {
	return validation.ValidateStruct(u,
		validation.Field(&u.PhoneNumber,
			validation.When(!isUpdate, validation.Required.Error("phone number is required")),
		),
		validation.Field(&u.FullName,
			validation.When(!isUpdate, validation.Required.Error("full name is required")),
			validation.Length(3, 100).Error("full name must be between 3 and 100 characters")),
		validation.Field(&u.MerchantID,
			validation.When(!isUpdate, validation.Required.Error("merchant id is required")),
		),
	)
}

func (u *UserLookUpRequest) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.PhoneNumber, validation.Required.Error("phone number is required")),
	)
}
