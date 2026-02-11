package auth

import (
	"lazeez-core/internal/common"
	"lazeez-core/internal/users"
	"regexp"
)

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

func (s *authService) generateTokens(user *users.User) (string, string, error) {
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
