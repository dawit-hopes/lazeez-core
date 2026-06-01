package key

import (
	"errors"
	"lazeez-core/config"
	"maps"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type KeyService interface {
	DecodeJWTToken(tokenStr string, secretKey string) (map[string]any, error)
	GenerateJWTToken(payload map[string]any, expirationMinutes int) (string, error)
	HashPassword(password string) (string, error)
	VerifyPassword(password string, hashedPassword string) (bool, error)
	ComparePassword(hashedPassword string, password string) error
}

type keyService struct {
	logger    config.Logger
	secretKey string
}

func NewKeyService(logger config.Logger, secretKey string) KeyService {
	return &keyService{
		logger:    logger,
		secretKey: secretKey,
	}
}

func (s *keyService) DecodeJWTToken(tokenStr string, secretKey string) (map[string]any, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.secretKey), nil
	})
	if err != nil {
		s.logger.Error("failed to parse token | error: " + err.Error())
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		s.logger.Error("invalid token claims")
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

func (s *keyService) GenerateJWTToken(payload map[string]any, expirationMinutes int) (string, error) {
	claims := jwt.MapClaims{}
	maps.Copy(claims, payload)
	if expirationMinutes > 0 {
		claims["exp"] = time.Now().Add(time.Minute * time.Duration(expirationMinutes)).Unix()
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		s.logger.Error("failed to sign token", "error", err)
		return "", err
	}
	return tokenStr, nil
}

func (s *keyService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", "error", err)
		return "", err
	}
	return string(hashedPassword), nil
}

func (s *keyService) VerifyPassword(password string, hashedPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		s.logger.Error("failed to verify password", "error", err)
		return false, err
	}
	return true, nil
}

func (s *keyService) ComparePassword(hashedPassword string, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		s.logger.Error("failed to compare password", "error", err)
		return err
	}
	return nil
}