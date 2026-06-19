package feedback

import (
	"context"

	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
)

// ListScope describes how staff feedback list/get requests are scoped.
type ListScope struct {
	BranchID   string
	MerchantID string
}

func ResolveListScope(ctx context.Context, queryBranchID string) (ListScope, error) {
	role, ok := middleware.GetRoleFromContext(ctx)
	if !ok {
		return ListScope{}, common.ErrUnAuthorized
	}

	switch role {
	case "super_admin":
		if queryBranchID != "" {
			return ListScope{BranchID: queryBranchID}, nil
		}
		return ListScope{}, nil
	case "super_branch_admin":
		merchantID, ok := middleware.GetMerchantIDFromContext(ctx)
		if !ok || merchantID == "" {
			return ListScope{}, common.ErrUnAuthorized
		}
		scope := ListScope{MerchantID: merchantID}
		if queryBranchID != "" {
			scope.BranchID = queryBranchID
		}
		return scope, nil
	case "branch_manager", "front_desk_agent", "room_service_staff":
		branchID, ok := middleware.GetBranchIDFromContext(ctx)
		if !ok || branchID == "" {
			return ListScope{}, common.ErrUnAuthorized
		}
		return ListScope{BranchID: branchID}, nil
	default:
		return ListScope{}, common.ErrUnAuthorized
	}
}
