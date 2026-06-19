package merchant

type SubscriptionPlan string

const (
	SubscriptionPlanDigitalMenu SubscriptionPlan = "DIGITAL_MENU"
	SubscriptionPlanOrdering    SubscriptionPlan = "ORDERING"
)

var validSubscriptionPlans = []any{
	SubscriptionPlanDigitalMenu,
	SubscriptionPlanOrdering,
}

// SubscriptionFeatures describes capabilities enabled for a plan.
// Feature flags are hardcoded per plan; only the plan name is stored on the merchant.
type SubscriptionFeatures struct {
	Menu     bool `json:"menu"`
	Ordering bool `json:"ordering"`
	Payments bool `json:"payments"`
}

func FeaturesForPlan(plan SubscriptionPlan) SubscriptionFeatures {
	switch plan {
	case SubscriptionPlanOrdering:
		return SubscriptionFeatures{Menu: true, Ordering: true, Payments: true}
	default:
		return SubscriptionFeatures{Menu: true, Ordering: false, Payments: false}
	}
}

func (p SubscriptionPlan) IsValid() bool {
	switch p {
	case SubscriptionPlanDigitalMenu, SubscriptionPlanOrdering:
		return true
	default:
		return false
	}
}
