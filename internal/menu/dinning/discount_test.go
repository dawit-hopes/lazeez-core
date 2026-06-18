package menu

import "testing"

func TestMenuDiscount_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		discount *MenuDiscount
		price    float64
		want     bool
	}{
		{
			name:     "nil discount",
			discount: nil,
			price:    100,
			want:     true,
		},
		{
			name:     "valid percentage",
			discount: &MenuDiscount{Type: DiscountTypePercentage, Value: 15},
			price:    100,
			want:     true,
		},
		{
			name:     "invalid percentage over 100",
			discount: &MenuDiscount{Type: DiscountTypePercentage, Value: 101},
			price:    100,
			want:     false,
		},
		{
			name:     "valid fixed",
			discount: &MenuDiscount{Type: DiscountTypeFixed, Value: 25},
			price:    100,
			want:     true,
		},
		{
			name:     "invalid fixed equal to price",
			discount: &MenuDiscount{Type: DiscountTypeFixed, Value: 100},
			price:    100,
			want:     false,
		},
		{
			name:     "invalid type",
			discount: &MenuDiscount{Type: "half_off", Value: 10},
			price:    100,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.discount.IsValid(tt.price); got != tt.want {
				t.Fatalf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseMenuDiscount(t *testing.T) {
	discount, err := parseMenuDiscount(`{"type":"percentage","value":10}`)
	if err != nil {
		t.Fatalf("parseMenuDiscount() error = %v", err)
	}
	if discount == nil || discount.Type != DiscountTypePercentage || discount.Value != 10 {
		t.Fatalf("unexpected discount: %#v", discount)
	}

	cleared, err := parseMenuDiscount("null")
	if err != nil {
		t.Fatalf("parseMenuDiscount(null) error = %v", err)
	}
	if cleared != nil {
		t.Fatalf("expected nil discount, got %#v", cleared)
	}
}
