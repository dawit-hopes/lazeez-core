package middleware

import (
	"bytes"
	"context"
	"io"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"lazeez-core/internal/key"
	"lazeez-core/internal/session"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/cors"
)

type Middleware interface {
	ValidateToken(next http.Handler) http.Handler
	RequireSessionKey(next http.Handler) http.Handler
	RequireBranch(next http.Handler) http.Handler
	RequireSuperAdmin(next http.Handler) http.Handler
	RequireRoomManagement(next http.Handler) http.Handler
	RequireFrontDesk(next http.Handler) http.Handler
	RequireOrderListAccess(next http.Handler) http.Handler
	RequireFeedbackListAccess(next http.Handler) http.Handler
	NotFoundHandler(w http.ResponseWriter, r *http.Request)
	MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request)
	CORSHandler(next http.Handler) http.Handler
	ValidateWebhook(next http.Handler) http.Handler
}

type contextKey string

const claimsContextKey contextKey = "claims"
const ChapaSignatureContextKey contextKey = "chapaSignature"
const ChapaXSignatureContextKey contextKey = "chapaXSignature"
const WebhookBodyContextKey contextKey = "webhookBody"

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

// RequireSessionKey ensures X-Session-Key header or session_key query param is present.
func (m *middleware) RequireSessionKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Session-Key")
		if key == "" {
			key = r.URL.Query().Get("session_key")
		}
		if strings.TrimSpace(key) == "" {
			common.WriteErrorResponse(w, common.ErrUnAuthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireBranch ensures user has branch_id (branch_manager, front_desk_agent, room_service_staff). Use after ValidateToken.
func (m *middleware) RequireBranch(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		branchID, ok := GetBranchIDFromContext(r.Context())
		if !ok || branchID == "" {
			common.WriteErrorResponse(w, common.ErrUnAuthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireSuperAdmin ensures user has super_admin role. Use after ValidateToken.
func (m *middleware) RequireSuperAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := GetRoleFromContext(r.Context())
		if !ok || role != "super_admin" {
			common.WriteErrorResponse(w, common.ErrUnAuthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireRoomManagement allows managing physical rooms: super_admin, branch_manager,
// or front_desk_agent. super_branch_admin (read-only) and others are rejected. Use after ValidateToken.
func (m *middleware) RequireRoomManagement(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := GetRoleFromContext(r.Context())
		if !ok {
			common.WriteErrorResponse(w, common.ErrUnAuthorized)
			return
		}
		switch role {
		case "super_admin", "branch_manager", "front_desk_agent":
			next.ServeHTTP(w, r)
		default:
			common.WriteErrorResponse(w, common.ErrUnAuthorized)
		}
	})
}

// RequireFrontDesk allows only front_desk_agent to mutate bookings (check-in/check-out/cancel).
// Other roles (including branch_manager and super_admin) are rejected. Use after ValidateToken.
func (m *middleware) RequireFrontDesk(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := GetRoleFromContext(r.Context())
		if !ok || role != "front_desk_agent" {
			common.WriteErrorResponse(w, common.ErrUnAuthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireOrderListAccess allows super_admin (all orders) or super_branch_admin (merchant-scoped).
func (m *middleware) RequireOrderListAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := GetRoleFromContext(r.Context())
		if !ok {
			common.WriteErrorResponse(w, common.ErrUnAuthorized)
			return
		}
		switch role {
		case "super_admin":
			next.ServeHTTP(w, r)
		case "super_branch_admin":
			if merchantID, ok := GetMerchantIDFromContext(r.Context()); !ok || merchantID == "" {
				common.WriteErrorResponse(w, common.ErrUnAuthorized)
				return
			}
			next.ServeHTTP(w, r)
		default:
			common.WriteErrorResponse(w, common.ErrUnAuthorized)
		}
	})
}

// RequireFeedbackListAccess allows staff roles that can view feedback in command center:
// super_admin (all), super_branch_admin (merchant), branch staff (branch).
func (m *middleware) RequireFeedbackListAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := GetRoleFromContext(r.Context())
		if !ok {
			common.WriteErrorResponse(w, common.ErrUnAuthorized)
			return
		}
		switch role {
		case "super_admin":
			next.ServeHTTP(w, r)
		case "super_branch_admin":
			if merchantID, ok := GetMerchantIDFromContext(r.Context()); !ok || merchantID == "" {
				common.WriteErrorResponse(w, common.ErrUnAuthorized)
				return
			}
			next.ServeHTTP(w, r)
		case "branch_manager", "front_desk_agent", "room_service_staff":
			if branchID, ok := GetBranchIDFromContext(r.Context()); !ok || branchID == "" {
				common.WriteErrorResponse(w, common.ErrUnAuthorized)
				return
			}
			next.ServeHTTP(w, r)
		default:
			common.WriteErrorResponse(w, common.ErrUnAuthorized)
		}
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

func (m *middleware) validateSession(ctx context.Context, uid string, token string) error {
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

	merchantID := ""
	if midRaw, exists := claims["mid"]; exists && midRaw != nil {
		if midStr, ok := midRaw.(string); ok {
			merchantID = midStr
		}
	}

	return map[string]any{
		"uid": uid,
		"bid": branchID,
		"rol": roleStr,
		"mid": merchantID,
	}, nil
}

// Default CORS origins when CORS_ALLOWED_ORIGINS is not set (local dev).
var defaultCORSOrigins = []string{
	"http://localhost:5173",
	"http://127.0.0.1:5173",
	"http://localhost:8080",
	"http://127.0.0.1:8080",
	"http://localhost:8081",
	"http://127.0.0.1:8081",
	"http://localhost:8082",
	"http://127.0.0.1:8082",
	"http://localhost:3000",
	"http://127.0.0.1:3000",
	"http://10.121.241.209:8081",
	"http://172.21.0.1:8081",
	"http://172.19.0.1:8081",
	"http://172.23.0.1:8081",
	"http://172.24.0.1:8081",
	"http://10.22.209.1:8082",
	"http://172.18.0.1:8082",
}

func getAllowedOrigins() []string {
	raw := os.Getenv("CORS_ALLOWED_ORIGINS")
	if raw == "" {
		return defaultCORSOrigins
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if o := strings.TrimSpace(p); o != "" {
			origins = append(origins, o)
		}
	}

	if len(origins) == 0 {
		return defaultCORSOrigins
	}
	return origins
}

func (m *middleware) CORSHandler(next http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins: getAllowedOrigins(),
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders: []string{
			"Content-Type", "Authorization", "X-Requested-With", "X-CSRF-Token",
			"Origin", "Accept",
			"X-Session-Key",
			"X-Room-Reference",
			"X-Pass-Code",
		},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           86400,
	})(next)
}

// GetRoleFromContext extracts the role (\"rol\") claim from the request context.
// It returns the role string and a boolean indicating whether it was present.
func GetRoleFromContext(ctx context.Context) (string, bool) {
	claims, ok := ctx.Value(claimsContextKey).(map[string]any)
	if !ok || claims == nil {
		return "", false
	}
	role, ok := claims["rol"].(string)
	return role, ok
}

func GetBranchIDFromContext(ctx context.Context) (string, bool) {
	claims, ok := ctx.Value(claimsContextKey).(map[string]any)
	if !ok || claims == nil {
		return "", false
	}
	branchID, ok := claims["bid"].(string)
	return branchID, ok
}

// GetUserIDFromContext extracts the user ID ("uid") claim from the request context.
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	claims, ok := ctx.Value(claimsContextKey).(map[string]any)
	if !ok || claims == nil {
		return "", false
	}
	uid, ok := claims["uid"].(string)
	if !ok || uid == "" {
		return "", false
	}
	return uid, true
}

// GetMerchantIDFromContext extracts the merchant ID ("mid") claim from the request context.
func GetMerchantIDFromContext(ctx context.Context) (string, bool) {
	claims, ok := ctx.Value(claimsContextKey).(map[string]any)
	if !ok || claims == nil {
		return "", false
	}
	mid, ok := claims["mid"].(string)
	if !ok || mid == "" {
		return "", false
	}
	return mid, true
}

func (m *middleware) ValidateWebhook(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		signature := r.Header.Get("chapa-signature")
		xSignature := r.Header.Get("x-chapa-signature")

		if signature == "" && xSignature == "" {
			common.WriteErrorResponse(w, common.ErrUnAuthorized)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			m.logger.Error("failed to read webhook body", "error", err)
			common.WriteErrorResponse(w, common.ErrInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))

		ctx := context.WithValue(r.Context(), ChapaSignatureContextKey, signature)
		ctx = context.WithValue(ctx, ChapaXSignatureContextKey, xSignature)
		ctx = context.WithValue(ctx, WebhookBodyContextKey, body)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
