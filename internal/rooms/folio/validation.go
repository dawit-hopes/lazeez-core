package folio

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (r *SettleBillInput) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.PaymentMethod,
			validation.Required.Error("payment method is required"),
			validation.In("cash", "card", "mobile", "other").Error("payment method must be cash, card, mobile, or other"),
		),
	)
}
