package feedback

import (
	"testing"
	"time"
)

func TestRatingRequest_Validate(t *testing.T) {
	valid := RatingRequest{
		Rating:      4,
		Comment:     "Great meal",
		Tags:        []string{"fast", "friendly"},
		PhoneNumber: "+251911234567",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	missingRating := RatingRequest{Comment: "No score"}
	if err := missingRating.Validate(); err == nil {
		t.Fatal("expected validation error for missing rating")
	}

	lowRating := RatingRequest{Rating: 0}
	if err := lowRating.Validate(); err == nil {
		t.Fatal("expected validation error for rating below 1")
	}

	highRating := RatingRequest{Rating: 6}
	if err := highRating.Validate(); err == nil {
		t.Fatal("expected validation error for rating above 5")
	}

	longComment := RatingRequest{
		Rating:  3,
		Comment: string(make([]byte, 501)),
	}
	if err := longComment.Validate(); err == nil {
		t.Fatal("expected validation error for comment longer than 500 characters")
	}

	longPhone := RatingRequest{
		Rating:      3,
		PhoneNumber: string(make([]byte, 21)),
	}
	if err := longPhone.Validate(); err == nil {
		t.Fatal("expected validation error for phone_number longer than 20 characters")
	}

	emptyTag := RatingRequest{
		Rating: 3,
		Tags:   []string{"good", ""},
	}
	if err := emptyTag.Validate(); err == nil {
		t.Fatal("expected validation error for empty tag")
	}
}

func TestStayRatingRequest_Validate(t *testing.T) {
	valid := StayRatingRequest{
		Rating:    5,
		TableName: "Table 12",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	longTableName := StayRatingRequest{
		Rating:    4,
		TableName: string(make([]byte, 256)),
	}
	if err := longTableName.Validate(); err == nil {
		t.Fatal("expected validation error for table_name longer than 255 characters")
	}
}

func TestValidateRatingDateRange(t *testing.T) {
	from := parseTestDate(t, "2026-06-01")
	to := parseTestDate(t, "2026-06-10")
	filter := RatingFilter{DateFrom: &from, DateTo: &to}
	if err := ValidateRatingDateRange(filter); err != nil {
		t.Fatalf("ValidateRatingDateRange() error = %v", err)
	}

	badTo := parseTestDate(t, "2026-05-01")
	badFilter := RatingFilter{DateFrom: &from, DateTo: &badTo}
	if err := ValidateRatingDateRange(badFilter); err == nil {
		t.Fatal("expected validation error when date_from is after date_to")
	}
}

func parseTestDate(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		t.Fatalf("failed to parse test date: %v", err)
	}
	return parsed
}
