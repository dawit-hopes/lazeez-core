package merchant

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (m *CreateMerchantRequest) Validate() error {
	return validation.ValidateStruct(m,
		validation.Field(&m.Name, validation.Required.Error("name is required"), validation.Length(3, 100).Error("name must be between 3 and 100 characters")),
	)
}

func (m *UpdateMerchantRequest) Validate() error {
	return validation.ValidateStruct(m,
		validation.Field(&m.Name, validation.Required.Error("name is required"), validation.Length(3, 100).Error("name must be between 3 and 100 characters")),
	)
}
