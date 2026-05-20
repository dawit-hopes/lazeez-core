package option

import "testing"

func TestModifierOptionRequest_Validate_priceAdjustmentFloat(t *testing.T) {
	req := ModifierOptionRequest{
		Name:            "Large",
		PriceAdjustment: 10.5,
		IsDefault:       false,
		IsAvailable:     true,
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("expected valid option, got %v", err)
	}
}
