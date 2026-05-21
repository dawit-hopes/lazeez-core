package clientsession

import (
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (r *CreateSessionInput) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.Reference,
			validation.Required.Error("reference is required"),
			validation.By(func(value any) error {
				ref, _ := value.(string)
				if strings.TrimSpace(ref) == "" {
					return validation.NewError("validation_reference", "reference is required")
				}
				return nil
			}),
		),
	)
}
