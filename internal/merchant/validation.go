package merchant

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (m *MerchantRequest) Validate() error {
	return validation.ValidateStruct(m,
		validation.Field(&m.Name, validation.Required.Error("name is required"), validation.Length(3, 100).Error("name must be between 3 and 100 characters")),
		validation.Field(&m.Logo, validation.Required.Error("logo is required")),
		validation.Field(&m.LogoHeader, validation.Required.Error("logo header is required")),
		validation.Field(&m.BranchType, validation.Required.Error("branch type is required")),
		validation.Field(&m.BranchType, validation.In(BranchTypeRestaurant, BranchTypeHotel).Error("branch type must be restaurant or hotel")),
		validation.Field(&m.SubscriptionPlan, validation.Required.Error("subscription plan is required")),
		validation.Field(&m.SubscriptionPlan, validation.In(validSubscriptionPlans...).Error("subscription plan must be DIGITAL_MENU or ORDERING")),
	)
}

func (m *MerchantRequest) ValidateUpdate() error {
	if m.SubscriptionPlan != "" {
		if err := validation.Validate(m.SubscriptionPlan, validation.In(validSubscriptionPlans...).Error("subscription plan must be DIGITAL_MENU or ORDERING")); err != nil {
			return err
		}
	}
	return nil
}

func IsEmpty(m *MerchantRequest) bool {
	return m.Name == "" && m.Logo == nil && m.BranchType == "" && m.SubscriptionPlan == ""
}
