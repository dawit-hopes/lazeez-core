package users

import (
	"context"
	"lazeez-core/internal/common"
	"regexp"
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
	user.FullName = req.FullName
	user.PhoneNumber = req.PhoneNumber
	user.BranchID = common.ToNUllString(req.BranchID)
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
		user.FullName = req.FullName
	}

	return user
}

func (s *userService) validatePassword(password string) error {
	if password == "" {
		s.logger.Error("Password is required; received empty value")
		return common.ErrPasswordRequired
	}

	// Length check: minimum 9 characters
	if len(password) < 9 {
		s.logger.Error("Password is too short; must be at least 9 characters long")
		return common.ErrPasswordTooShort
	}

	// Must contain at least one letter (a–z or A–Z)
	letterRegex := regexp.MustCompile(`[A-Za-z]`)
	if !letterRegex.MatchString(password) {
		s.logger.Error("Password validation failed: missing letter (a-z or A-Z)")
		return common.ErrPasswordMissingLetter
	}

	// Must contain at least one number (0–9) OR one special character
	digitRegex := regexp.MustCompile(`[0-9]`)
	specialCharRegex := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};:'",.<>/?|\\]`)
	if !(digitRegex.MatchString(password) || specialCharRegex.MatchString(password)) {
		s.logger.Error("Password validation failed: missing number (0-9) or special character")
		return common.ErrPasswordMissingNumberOrSpecial
	}

	return nil
}
