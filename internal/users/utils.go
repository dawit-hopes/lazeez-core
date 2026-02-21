package users

import (
	"context"
	"lazeez-core/internal/common"
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

	return user
}
