package promotion

import (
	"bytes"
	"testing"

	"lazeez-core/internal/common"
)

func TestParseFormDate(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{name: "iso date", raw: "2026-06-18", want: "2026-06-18"},
		{name: "us date", raw: "06/18/2026", want: "2026-06-18"},
		{name: "empty", raw: "", want: ""},
		{name: "invalid", raw: "18-06-2026", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFormDate(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseFormDate() error = %v", err)
			}
			if tt.want == "" {
				if !got.IsZero() {
					t.Fatalf("expected zero date, got %v", got.Time())
				}
				return
			}
			if got.Time().Format("2006-01-02") != tt.want {
				t.Fatalf("parseFormDate() = %v, want %v", got.Time().Format("2006-01-02"), tt.want)
			}
		})
	}
}

func TestPromotionRequest_Validate(t *testing.T) {
	start := common.Date{}
	_ = start.UnmarshalJSON([]byte(`"2026-06-18"`))
	end := common.Date{}
	_ = end.UnmarshalJSON([]byte(`"2026-07-18"`))

	req := PromotionRequest{
		Title:     "Weekend Feast",
		StartDate: start,
		EndDate:   end,
	}
	if err := req.Validate(true); err == nil {
		t.Fatal("expected create validation to require banner image")
	}

	req.BannerImage = &testMultipartFile{Reader: bytes.NewReader(nil)}
	if err := req.Validate(true); err != nil {
		t.Fatalf("Validate(create) error = %v", err)
	}

	updateReq := PromotionRequest{Title: "Updated"}
	if err := updateReq.Validate(false); err != nil {
		t.Fatalf("Validate(update) error = %v", err)
	}

	badRange := PromotionRequest{
		Title:     "Weekend Feast",
		StartDate: start,
		EndDate:   common.Date{},
		BannerImage: &testMultipartFile{Reader: bytes.NewReader(nil)},
	}
	_ = badRange.EndDate.UnmarshalJSON([]byte(`"2026-05-01"`))
	if err := badRange.Validate(true); err == nil {
		t.Fatal("expected validation error for end date before start date")
	}
}

type testMultipartFile struct {
	*bytes.Reader
}

func (f *testMultipartFile) Close() error { return nil }
