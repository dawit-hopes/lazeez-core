package check

import "lazeez-core/internal/users"

// deriveReadiness computes the readiness state from item counts and the merchant's policy.
//   - not_ready:          any item still 'sent'
//   - ready_for_billing:  strict = all items 'ready'; lenient = no item 'sent'
//   - in_progress:        otherwise
//
// A check with no billable items is never ready (there is nothing to bill yet).
func deriveReadiness(counts ReadinessCounts, policy string) ReadinessState {
	if counts.TotalItems == 0 {
		return ReadinessNotReady
	}
	if counts.Sent > 0 {
		return ReadinessNotReady
	}
	// No 'sent' items remain at this point.
	if policy == BillPrintPolicyLenient {
		return ReadinessReadyForBilling
	}
	if counts.Ready == counts.TotalItems {
		return ReadinessReadyForBilling
	}
	return ReadinessInProgress
}

// canSettleChecks reports roles allowed to settle a table check: cashiers and branch managers.
func canSettleChecks(role string) bool {
	switch users.Role(role) {
	case users.RoleCashier, users.RoleBranchManager:
		return true
	default:
		return false
	}
}
