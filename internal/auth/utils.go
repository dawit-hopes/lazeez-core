package auth

import (
	"database/sql"
	"lazeez-core/internal/common"
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

func (s *authService) generateTokens(user map[string]any) (string, string, error) {
	branchID := ""
	if bid := user["bid"]; bid != nil {
		switch v := bid.(type) {
		case string:
			branchID = v
		case sql.NullString:
			if v.Valid {
				branchID = v.String
			}
		}
	}
	payload := map[string]any{"uid": user["id"], "bid": branchID, "rol": user["role"]}

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
