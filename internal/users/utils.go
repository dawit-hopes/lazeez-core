package users

import (
	"context"
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"time"

	"github.com/google/uuid"
)

func (s *userService) validateBranch(ctx context.Context, branchID, merchantID string) error {
	branchRes, err := s.branchService.Get(ctx, branchID)
	if err != nil {
		s.logger.Error("Failed to get branch", "error", err)
		return err
	}
	if branchRes.ID == "" {
		return common.ErrBranchNotFound
	}

	if branchRes.MerchantID != merchantID {
		return common.ErrBranchNotFound
	}
	return nil
}

func (s *userService) validateCreateSuperAdminAccess(ctx context.Context) error {
	role, ok := middleware.GetRoleFromContext(ctx)
	if !ok || Role(role) != RoleAdmin {
		return common.ErrUnAuthorized
	}
	return nil
}

func (s *userService) validateCreateUserAccess(ctx context.Context, req UserRequest) error {
	roleStr, ok := middleware.GetRoleFromContext(ctx)
	if !ok {
		return common.ErrUnAuthorized
	}
	creatorRole := Role(roleStr)

	_, branchType, err := s.userRepository.ResolveMerchantContext(ctx, req.BranchID, req.MerchantID)
	if err != nil {
		s.logger.Error("Failed to resolve merchant context for user creation", "error", err)
		return err
	}

	if !CanCreatorAssignRole(creatorRole, req.Role, BranchType(branchType)) {
		return common.ErrRoleNotAllowedForCreation
	}

	switch creatorRole {
	case RoleSuperBranchManager:
		creatorMerchantID, ok := middleware.GetMerchantIDFromContext(ctx)
		if !ok || creatorMerchantID == "" || creatorMerchantID != req.MerchantID {
			return common.ErrUnAuthorized
		}
	case RoleBranchManager:
		creatorBranchID, ok := middleware.GetBranchIDFromContext(ctx)
		if !ok || creatorBranchID == "" || creatorBranchID != req.BranchID {
			return common.ErrUnAuthorized
		}
		creatorMerchantID, ok := middleware.GetMerchantIDFromContext(ctx)
		if !ok || creatorMerchantID == "" || creatorMerchantID != req.MerchantID {
			return common.ErrUnAuthorized
		}
	default:
		return common.ErrUnAuthorized
	}

	return nil
}

type viewerContext struct {
	role       Role
	merchantID string
	branchID   string
	branchType BranchType
}

func (s *userService) resolveViewerContext(ctx context.Context) (viewerContext, error) {
	roleStr, ok := middleware.GetRoleFromContext(ctx)
	if !ok {
		return viewerContext{}, common.ErrUnAuthorized
	}
	vc := viewerContext{
		role:       Role(roleStr),
		merchantID: "",
		branchID:   "",
	}
	if mid, ok := middleware.GetMerchantIDFromContext(ctx); ok {
		vc.merchantID = mid
	}
	if bid, ok := middleware.GetBranchIDFromContext(ctx); ok {
		vc.branchID = bid
	}
	if vc.role == RoleBranchManager {
		_, branchType, err := s.userRepository.ResolveMerchantContext(ctx, vc.branchID, vc.merchantID)
		if err != nil {
			return viewerContext{}, err
		}
		vc.branchType = BranchType(branchType)
	}
	return vc, nil
}

func (s *userService) userMerchantID(ctx context.Context, user *User) (string, error) {
	if user.MerchantID.Valid && user.MerchantID.String != "" {
		return user.MerchantID.String, nil
	}
	if user.BranchID.Valid && user.BranchID.String != "" {
		merchantID, _, err := s.userRepository.ResolveMerchantContext(ctx, user.BranchID.String, "")
		return merchantID, err
	}
	return "", nil
}

func (s *userService) validateUserViewAccess(ctx context.Context, user *User) error {
	vc, err := s.resolveViewerContext(ctx)
	if err != nil {
		return err
	}

	if !CanViewerSeeUser(vc.role, user.Role, vc.branchType) {
		return common.ErrUserNotFound
	}

	switch vc.role {
	case RoleAdmin:
		return nil
	case RoleSuperBranchManager:
		if vc.merchantID == "" {
			return common.ErrUnAuthorized
		}
		targetMerchantID, err := s.userMerchantID(ctx, user)
		if err != nil {
			return err
		}
		if targetMerchantID != vc.merchantID {
			return common.ErrUserNotFound
		}
		return nil
	case RoleBranchManager:
		if vc.branchID == "" || !user.BranchID.Valid || user.BranchID.String != vc.branchID {
			return common.ErrUserNotFound
		}
		return nil
	default:
		return common.ErrUnAuthorized
	}
}

func (s *userService) createUserDefaultData(req *UserRequest) *User {
	var user User
	now := time.Now()
	user.ID = uuid.New().String()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.IsFirstLogin = true
	user.IsLocked = false
	user.LoggingAttempts = 0
	user.Role = RoleBranchManager
	user.FullName = common.FormatText(req.FullName)
	user.PhoneNumber = req.PhoneNumber
	user.BranchID = common.ToNUllString(req.BranchID)
	user.MerchantID = common.ToNUllString(req.MerchantID)
	user.Role = req.Role
	return &user
}

func (s *userService) createSuperAdminUserDefaultData(req *SuperAdminUserRequest) *User {
	var user User
	now := time.Now()
	user.ID = uuid.New().String()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.IsFirstLogin = true
	user.IsLocked = false
	user.LoggingAttempts = 0
	user.Role = RoleSuperBranchManager
	user.FullName = common.FormatText(req.FullName)
	user.PhoneNumber = req.PhoneNumber
	user.MerchantID = common.ToNUllString(req.MerchantID)
	user.Role = RoleSuperBranchManager
	return &user
}

func (s *userService) updateUserDefaultData(req *UserRequest, existingUser *User) *User {
	user := existingUser
	now := time.Now()
	user.UpdatedAt = now
	if req.PhoneNumber != "" {
		user.PhoneNumber = req.PhoneNumber
	}
	if req.BranchID != "" {
		user.BranchID = common.ToNUllString(req.BranchID)
	}
	if req.FullName != "" {
		user.FullName = common.FormatText(req.FullName)
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	return user
}
