package promotion

import (
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewPromotionRoutes(router chi.Router, handler PromotionHandler, mw middleware.Middleware) {
	routes := []common.Route{
		{
			Method:      http.MethodPost,
			Path:        "/promotions",
			Handler:     handler.Create,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
		{
			Method:      http.MethodGet,
			Path:        "/promotions",
			Handler:     handler.List,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
		{
			Method:      http.MethodGet,
			Path:        "/promotions/{id}",
			Handler:     handler.Get,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
		{
			Method:      http.MethodPut,
			Path:        "/promotions/{id}",
			Handler:     handler.Update,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
		{
			Method:      http.MethodDelete,
			Path:        "/promotions/{id}",
			Handler:     handler.Delete,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
		{
			Method:      http.MethodPost,
			Path:        "/promotions/{id}/undelete",
			Handler:     handler.UnDelete,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
	}
	common.RegisterRoutes(router, routes)
}
