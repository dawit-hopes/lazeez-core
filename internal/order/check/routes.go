package check

import (
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewCheckRoutes(router chi.Router, handler CheckHandler, mw middleware.Middleware) {
	cashierMw := []func(next http.Handler) http.Handler{mw.ValidateToken, mw.RequireCashier}

	routes := []common.Route{
		{
			Method:      http.MethodGet,
			Path:        "/cashier/checks",
			Handler:     handler.ListOpen,
			Middlewares: cashierMw,
		},
		{
			Method:      http.MethodGet,
			Path:        "/cashier/checks/{id}",
			Handler:     handler.Get,
			Middlewares: cashierMw,
		},
		{
			Method:      http.MethodPut,
			Path:        "/cashier/checks/{id}/settle",
			Handler:     handler.Settle,
			Middlewares: cashierMw,
		},
	}
	common.RegisterRoutes(router, routes)
}
