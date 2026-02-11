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
			Path:    "/login",
			Handler: handler.Login,
		},
		{
			Method:  http.MethodPost,
			Path:    "/first-time-login",
			Handler: handler.FirstTimeLogin,
		},
		{
			Method:      http.MethodPost,
			Path:        "/logout/{id}",
			Handler:     handler.Logout,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
		{
			Method:  http.MethodPost,
			Path:    "/reset-password",
			Handler: handler.ResetPassword,
		},
		{
			Method:  http.MethodPost,
			Path:    "/refresh-token",
			Handler: handler.RefreshToken,
		},
	}
	common.RegisterRoutes(router, routes)
}
