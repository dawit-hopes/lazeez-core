package auth

import (
	"context"
	"lazeez-core/internal/common"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *authService) checkUserExistsByPhoneNumber(ctx context.Context, phoneNumber string) (*User, error) {
	existingUser, err := s.authRepository.GetUserByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		s.logger.Error("Failed to get user by phone number", "error", err)
		return nil, err
	}
	return &existingUser, nil
}

func (s *authService) normalizePhoneNumber(phoneNumber string) string {
	nonNumericRegex := regexp.MustCompile(`[^0-9]`)
	// 1. Strip everything except digits in one pass
	clean := nonNumericRegex.ReplaceAllString(phoneNumber, "")

	// 2. Handle empty strings to prevent index panics
	if clean == "" {
		return ""
	}

	// 3. Normalize local formats to 251
	if strings.HasPrefix(clean, "0") {
		return "251" + clean[1:]
	}

	// 4. If they start with 9 or 7, assume they forgot the prefix
	if len(clean) == 9 && (clean[0] == '9' || clean[0] == '7') {
		return "251" + clean
	}

	return clean
}

func (s *authService) validatePhoneNumber(phoneNumber string) (string, error) {
	s.logger.Info("Validating phone number", "phone number", phoneNumber)

	normalizedPhoneNumber := s.normalizePhoneNumber(phoneNumber)
	if normalizedPhoneNumber == "" {
		return "", common.ErrPhoneNumberRequired
	}

	re := regexp.MustCompile(`^251[79]\d{8}$`)
	if !re.MatchString(normalizedPhoneNumber) {
		return "", common.ErrInvalidPhoneNumber
	}
	return normalizedPhoneNumber, nil
}

func (s *authService) validateBranch(ctx context.Context, branchID, merchantID string) error {
	branch, err := s.branchService.Get(ctx, branchID)
	if err != nil {
		s.logger.Error("Failed to get branch", "error", err)
		return err
	}
	if branch.ID == "" {
		return common.ErrBranchNotFound
	}

	if branch.MerchantID != merchantID {
		return common.ErrBranchNotFound
	}
	return nil
}

func (s *authService) createUserDefaultData(req *UserRequest) *User {
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

func (s *authService) updateUserDefaultData(req *UserRequest, existingUser *User) *User {
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

func (s *authService) validatePassword(password string) error {
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

func (s *authService) generateTokens(user *User) (string, string, error) {
	// Normalize BranchID to a plain string for JWT payload
	branchID := ""
	if user.BranchID.Valid {
		branchID = user.BranchID.String
	}
	payload := map[string]any{"uid": user.ID, "bid": branchID, "rol": user.Role}

	accessToken, err := s.keyService.GenerateJWTToken(payload, accessTokenExpirationMinutes)
	if err != nil {
		s.logger.Error("Failed to generate access token", "error", err)
		return "", "", err
	}
	refreshToken, err := s.keyService.GenerateJWTToken(payload, refreshTokenExpirationMinutes)
	if err != nil {
		s.logger.Error("Failed to generate refresh token", "error", err)
		return "", "", err
	}
	return accessToken, refreshToken, nil
}
