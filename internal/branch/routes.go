package branch

import (
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewBranchRoutes(router chi.Router, handler BranchHandler, middleware middleware.Middleware) {
	routes := []common.Route{
		{
			Method:      http.MethodPost,
			Path:        "/branches",
			Handler:     handler.Create,
			Middlewares: []func(next http.Handler) http.Handler{middleware.ValidateToken},
		},
		{
			Method:      http.MethodGet,
			Path:        "/branches/{id}",
			Handler:     handler.Get,
			Middlewares: []func(next http.Handler) http.Handler{middleware.ValidateToken},
		},
		{
			Method:      http.MethodPatch,
			Path:        "/branches/{id}",
			Handler:     handler.Update,
			Middlewares: []func(next http.Handler) http.Handler{middleware.ValidateToken},
		},
		{
			Method:      http.MethodDelete,
			Path:        "/branches/{id}",
			Handler:     handler.Delete,
			Middlewares: []func(next http.Handler) http.Handler{middleware.ValidateToken},
		},
		{
			Method:      http.MethodGet,
			Path:        "/branches",
			Handler:     handler.GetAll,
			Middlewares: []func(next http.Handler) http.Handler{middleware.ValidateToken},
		},
		{
			Method:      http.MethodGet,
			Path:        "/branches/merchant/{merchantID}",
			Handler:     handler.GetAllByMerchantID,
			Middlewares: []func(next http.Handler) http.Handler{middleware.ValidateToken},
		},
		{
			Method:      http.MethodPost,
			Path:        "/branches/{id}/undelete",
			Handler:     handler.UnDelete,
			Middlewares: []func(next http.Handler) http.Handler{middleware.ValidateToken},
		},
	}
	common.RegisterRoutes(router, routes)
}
