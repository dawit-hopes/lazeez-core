package check

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (r *SettleCheckInput) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.PaymentMethod,
			validation.Required.Error("payment method is required"),
			validation.In("cash").Error("payment method must be cash"),
		),
	)
}
