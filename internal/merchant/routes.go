package merchant

import (
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewMerchantRoutes(router chi.Router, handler MerchantHandler, middleware middleware.Middleware) {
	routes := []common.Route{
		{
			Method:      http.MethodPost,
			Path:        "/merchants",
			Handler:     handler.Create,
			Middlewares: []func(next http.Handler) http.Handler{middleware.ValidateToken},
		},
		{
			Method:      http.MethodGet,
			Path:        "/merchants/{id}",
			Handler:     handler.Get,
			Middlewares: []func(next http.Handler) http.Handler{middleware.ValidateToken},
		},
		{
			Method:      http.MethodPatch,
			Path:        "/merchants/{id}",
			Handler:     handler.Update,
			Middlewares: []func(next http.Handler) http.Handler{middleware.ValidateToken},
		},
		{
			Method:      http.MethodDelete,
			Path:        "/merchants/{id}",
			Handler:     handler.Delete,
			Middlewares: []func(next http.Handler) http.Handler{middleware.ValidateToken},
		},
		{
			Method:      http.MethodGet,
			Path:        "/merchants",
			Handler:     handler.GetAll,
			Middlewares: []func(next http.Handler) http.Handler{middleware.ValidateToken},
		},
		{
			Method:      http.MethodPost,
			Path:        "/merchants/{id}/undelete",
			Handler:     handler.UnDelete,
			Middlewares: []func(next http.Handler) http.Handler{middleware.ValidateToken},
		},
	}
	common.RegisterRoutes(router, routes)
}
