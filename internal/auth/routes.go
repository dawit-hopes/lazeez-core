package auth

import (
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewAuthRoutes(router chi.Router, handler AuthHandler, mw middleware.Middleware) {
	routes := []common.Route{
		{
			Method:  http.MethodPost,
			Path:    "/auth/login",
			Handler: handler.Login,
		},
		{
			Method:  http.MethodPost,
			Path:    "/auth/first-time-login",
			Handler: handler.FirstTimeLogin,
		},
		{
			Method:      http.MethodPost,
			Path:        "/auth/logout/{id}",
			Handler:     handler.Logout,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
		{
			Method:  http.MethodPost,
			Path:    "/auth/reset-password",
			Handler: handler.ResetPassword,
		},
		{
			Method:  http.MethodPost,
			Path:    "/auth/refresh",
			Handler: handler.RefreshToken,
		},
	}
	common.RegisterRoutes(router, routes)
}
