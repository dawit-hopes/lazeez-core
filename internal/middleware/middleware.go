package middleware

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"lazeez-core/internal/key"
	"lazeez-core/internal/session"
	"net/http"
	"os"
	"strings"
)

type Middleware interface {
	ValidateToken(next http.Handler) http.Handler
	CORSHandler(next http.Handler) http.Handler
	NotFoundHandler(w http.ResponseWriter, r *http.Request)
	MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request)
}

type contextKey string

const claimsContextKey contextKey = "claims"

type middleware struct {
	keyService     key.KeyService
	sessionService session.SessionService
	logger         config.Logger
}

func NewMiddleware(keyService key.KeyService, sessionService session.SessionService, logger config.Logger) Middleware {
	return &middleware{keyService: keyService, sessionService: sessionService, logger: logger}
}

func (m *middleware) ValidateToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			common.WriteErrorResponse(w, common.ErrUnAuthorized)
			return
		}

		bearerToken := strings.Split(token, " ")
		if len(bearerToken) != 2 {
			common.WriteErrorResponse(w, common.ErrUnAuthorized)

			return
		}

		token = bearerToken[1]
		claim, err := m.decodeToken(token)
		if err != nil {
			common.WriteErrorResponse(w, err)
			return
		}

		sessionErr := m.validateSession(r.Context(), claim["uid"].(string), token)
		if sessionErr != nil {
			common.WriteErrorResponse(w, sessionErr)
			return
		}

		ctx := context.WithValue(r.Context(), claimsContextKey, claim)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// NotFoundHandler is intended to be registered with router.NotFound(...)
// and therefore does not call next.
func (m *middleware) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	resp := common.NewResponse(nil, "route not found", common.NotFound)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	if data := resp.ToJSON(); data != nil {
		_, _ = w.Write(data)
	}
}

// MethodNotAllowedHandler is intended to be registered with router.MethodNotAllowed(...)
// and therefore does not call next.
func (m *middleware) MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	resp := common.NewResponse(nil, "method not allowed", http.StatusMethodNotAllowed)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	if data := resp.ToJSON(); data != nil {
		_, _ = w.Write(data)
	}
}

func (m *middleware) CORSHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		next.ServeHTTP(w, r)
	})
}

func (m *middleware) validateSession(ctx context.Context, uid string, token string) error {
	m.logger.Info("validating session", "uid", uid)
	m.logger.Info("**********************************")
	session, err := m.sessionService.GetSession(ctx, uid)
	if err != nil {
		m.logger.Error("failed to get session", "error", err)
		return common.ErrUnAuthorized
	}

	if session.IsRevoked {
		m.logger.Error("session is revoked")
		return common.ErrUnAuthorized
	}

	if session.AccessToken != token {
		m.logger.Error("access token is invalid")
		return common.ErrUnAuthorized
	}

	if session.UserID != uid {
		m.logger.Error("user id is invalid")
		return common.ErrUnAuthorized
	}
	m.logger.Info("session validated successfully")
	return nil
}

func (m *middleware) decodeToken(token string) (map[string]any, error) {

	claims, err := m.keyService.DecodeJWTToken(token, os.Getenv("JWT_SECRET_KEY"))
	if err != nil {
		return nil, common.ErrUnAuthorized
	}

	uid, ok := claims["uid"].(string)
	if !ok {
		return nil, common.ErrUnAuthorized
	}

	// BranchID may be empty; ensure we safely coerce to string
	branchID := ""
	if bidRaw, exists := claims["bid"]; exists && bidRaw != nil {
		if bidStr, ok := bidRaw.(string); ok {
			branchID = bidStr
		}
	}

	roleStr, ok := claims["rol"].(string)
	if !ok {
		return nil, common.ErrUnAuthorized
	}

	return map[string]any{
		"uid": uid,
		"bid": branchID,
		"rol": roleStr,
	}, nil
}
