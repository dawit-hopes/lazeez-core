package feedback

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (r *RatingRequest) Validate() error {
	return validateRatingFields(r.Rating, r.Comment, r.PhoneNumber, r.Tags)
}

func (r *StayRatingRequest) Validate() error {
	if err := validateRatingFields(r.Rating, r.Comment, r.PhoneNumber, r.Tags); err != nil {
		return err
	}
	if r.TableName != "" {
		return validation.ValidateStruct(r,
			validation.Field(&r.TableName, validation.Length(0, 255).Error("table_name must be at most 255 characters")),
		)
	}
	return nil
}

func validateRatingFields(rating int, comment, phoneNumber string, tags []string) error {
	if err := validation.Validate(rating,
		validation.Required.Error("rating is required"),
		validation.Min(1).Error("rating must be between 1 and 5"),
		validation.Max(5).Error("rating must be between 1 and 5"),
	); err != nil {
		return err
	}
	if comment != "" {
		if err := validation.Validate(comment, validation.Length(0, 500).Error("comment must be at most 500 characters")); err != nil {
			return err
		}
	}
	if phoneNumber != "" {
		if err := validation.Validate(phoneNumber, validation.Length(0, 20).Error("phone_number must be at most 20 characters")); err != nil {
			return err
		}
	}
	for _, tag := range tags {
		tag = trimTag(tag)
		if tag == "" {
			return validation.Errors{"tags": validation.NewError("validation", "tags must not contain empty values")}.Filter()
		}
		if len(tag) > 100 {
			return validation.Errors{"tags": validation.NewError("validation", "each tag must be at most 100 characters")}.Filter()
		}
	}
	return nil
}

func trimTag(value string) string {
	for len(value) > 0 && (value[0] == ' ' || value[0] == '\t') {
		value = value[1:]
	}
	for len(value) > 0 && (value[len(value)-1] == ' ' || value[len(value)-1] == '\t') {
		value = value[:len(value)-1]
	}
	return value
}
