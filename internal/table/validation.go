package table

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (r *TableRequest) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.BranchID,
			validation.Required.Error("branch id is required"),
		),
		validation.Field(&r.Tables,
			validation.Required.Error("at least one table is required"),
			validation.Length(1, 100).Error("tables must contain between 1 and 100 items"),
			validation.Each(validation.By(func(v interface{}) error {
				t, ok := v.(TableQR)
				if !ok {
					return nil
				}
				return validation.ValidateStruct(&t,
					validation.Field(&t.TableName,
						validation.Required.Error("table name is required"),
					),
				)
			})),
		),
	)
}
