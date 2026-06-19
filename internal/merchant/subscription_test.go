package merchant

import "testing"

func TestFeaturesForPlan(t *testing.T) {
	tests := []struct {
		plan     SubscriptionPlan
		menu     bool
		ordering bool
		payments bool
	}{
		{
			plan: SubscriptionPlanDigitalMenu,
			menu: true, ordering: false, payments: false,
		},
		{
			plan: SubscriptionPlanOrdering,
			menu: true, ordering: true, payments: true,
		},
		{
			plan: SubscriptionPlan("UNKNOWN"),
			menu: true, ordering: false, payments: false,
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.plan), func(t *testing.T) {
			features := FeaturesForPlan(tt.plan)
			if features.Menu != tt.menu || features.Ordering != tt.ordering || features.Payments != tt.payments {
				t.Fatalf("FeaturesForPlan(%q) = %+v, want menu=%v ordering=%v payments=%v",
					tt.plan, features, tt.menu, tt.ordering, tt.payments)
			}
		})
	}
}

func TestSubscriptionPlan_IsValid(t *testing.T) {
	if !SubscriptionPlanDigitalMenu.IsValid() || !SubscriptionPlanOrdering.IsValid() {
		t.Fatal("expected known plans to be valid")
	}
	if SubscriptionPlan("INVALID").IsValid() {
		t.Fatal("expected unknown plan to be invalid")
	}
}

func TestMerchantRequest_ValidateSubscriptionPlan(t *testing.T) {
	if err := validationOnlySubscriptionPlan(SubscriptionPlanOrdering); err != nil {
		t.Fatalf("valid plan: %v", err)
	}
	if err := validationOnlySubscriptionPlan(SubscriptionPlan("BAD")); err == nil {
		t.Fatal("expected invalid subscription plan error")
	}
}

func validationOnlySubscriptionPlan(plan SubscriptionPlan) error {
	req := MerchantRequest{SubscriptionPlan: plan}
	return req.ValidateUpdate()
}
