package category

import (
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewCategoryRoutes(router chi.Router, handler CategoryHandler, mw middleware.Middleware) {
	routes := []common.Route{
		{
			Method:      http.MethodPost,
			Path:        "/categories",
			Handler:     handler.Create,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
		{
			Method:      http.MethodGet,
			Path:        "/categories/{id}",
			Handler:     handler.Get,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
		{
			Method:      http.MethodPut,
			Path:        "/categories/{id}",
			Handler:     handler.Update,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
		{
			Method:      http.MethodDelete,
			Path:        "/categories/{id}",
			Handler:     handler.Delete,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
		{
			Method:      http.MethodGet,
			Path:        "/categories",
			Handler:     handler.List,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
		{
			Method:      http.MethodPost,
			Path:        "/categories/{id}/undelete",
			Handler:     handler.UnDelete,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
	}
	common.RegisterRoutes(router, routes)
}
