package promotion

import (
	"lazeez-core/internal/common"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (r *PromotionRequest) Validate(isCreate bool) error {
	rules := []*validation.FieldRules{}

	if isCreate {
		rules = append(rules,
			validation.Field(&r.Title,
				validation.Required.Error("title is required"),
				validation.Length(3, 100).Error("title must be between 3 and 100 characters"),
			),
			validation.Field(&r.BannerImage,
				validation.Required.Error("banner image is required"),
			),
			validation.Field(&r.StartDate,
				validation.Required.Error("start date is required"),
			),
			validation.Field(&r.EndDate,
				validation.Required.Error("end date is required"),
			),
		)
	} else {
		if r.IsEmpty() {
			return validation.NewError("validation", "at least one field is required")
		}
		if r.Title != "" {
			rules = append(rules, validation.Field(&r.Title,
				validation.Length(3, 100).Error("title must be between 3 and 100 characters"),
			))
		}
	}

	if r.Description != "" {
		rules = append(rules, validation.Field(&r.Description,
			validation.Length(0, 1000).Error("description must be at most 1000 characters"),
		))
	}

	if !r.StartDate.IsZero() || !r.EndDate.IsZero() {
		rules = append(rules, validation.Field(&r.EndDate,
			validation.By(func(value any) error {
				end, ok := value.(common.Date)
				if !ok {
					return nil
				}
				start := r.StartDate
				if start.IsZero() && isCreate {
					return nil
				}
				if !start.IsZero() && !end.IsZero() && end.Time().Before(start.Time()) {
					return validation.NewError("validation", "end date must be on or after start date")
				}
				return nil
			}),
		))
	}

	return validation.ValidateStruct(r, rules...)
}
