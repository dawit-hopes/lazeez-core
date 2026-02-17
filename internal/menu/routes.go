package menu

import (
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewMenuRoutes(router chi.Router, handler MenuHandler, middleware middleware.Middleware) {
	routes := []common.Route{
		{
			Method:      http.MethodPost,
			Path:        "/menus",
			Handler:     handler.Create,
			Middlewares: []func(next http.Handler) http.Handler{middleware.ValidateToken},
		},
		{
			Method:      http.MethodGet,
			Path:        "/menus/{id}",
			Handler:     handler.Get,
			Middlewares: []func(next http.Handler) http.Handler{middleware.ValidateToken},
		},
		{
			Method:      http.MethodPut,
			Path:        "/menus/{id}",
			Handler:     handler.Update,
			Middlewares: []func(next http.Handler) http.Handler{middleware.ValidateToken},
		},
		{
			Method:      http.MethodDelete,
			Path:        "/menus/{id}",
			Handler:     handler.Delete,
			Middlewares: []func(next http.Handler) http.Handler{middleware.ValidateToken},
		},
		{
			Method:      http.MethodPost,
			Path:        "/menus/{id}/undelete",
			Handler:     handler.UnDelete,
			Middlewares: []func(next http.Handler) http.Handler{middleware.ValidateToken},
		},
		{
			Method:      http.MethodGet,
			Path:        "/menus",
			Handler:     handler.List,
			Middlewares: []func(next http.Handler) http.Handler{middleware.ValidateToken},
		},
	}
	common.RegisterRoutes(router, routes)
}
