package branch

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (r *CreateBranchRequest) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.MerchantID,
			validation.Required.Error("merchant id is required"),
			validation.Match(regexp.MustCompile(`^[0-9a-f-]+$`)).Error("merchant id must be a valid UUID"),
		),
		validation.Field(&r.BranchName,
			validation.Required.Error("branch name is required"),
			validation.Length(3, 100).Error("branch name must be between 3 and 100 characters"),
		),
		validation.Field(&r.Address,
			validation.Required.Error("address is required"),
			validation.Length(3, 255).Error("address must be between 3 and 255 characters"),
		),
		validation.Field(&r.PhoneNumber,
			validation.Required.Error("phone number is required"),
			validation.Match(regexp.MustCompile(`^[0-9]+$`)).Error("phone number must be numeric"),
		),
	)
}

func (r *UpdateBranchRequest) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.ID,
			validation.Required.Error("id is required"),
			validation.Match(regexp.MustCompile(`^[0-9a-f-]+$`)).Error("id must be a valid UUID"),
		),
		validation.Field(&r.BranchName,
			validation.Length(3, 100).Error("branch name must be between 3 and 100 characters"),
		),
		validation.Field(&r.Address,
			validation.Length(3, 255).Error("address must be between 3 and 255 characters"),
		),
		validation.Field(&r.PhoneNumber,
			validation.Match(regexp.MustCompile(`^[0-9]+$`)).Error("phone number must be numeric"),
		),
	)
}
