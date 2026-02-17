package ingredient

import (
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewIngredientRoutes(router chi.Router, handler IngredientHandler, mw middleware.Middleware) {
	routes := []common.Route{
		{
			Method:      http.MethodPost,
			Path:        "/ingredients",
			Handler:     handler.Create,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
		{
			Method:      http.MethodGet,
			Path:        "/ingredients/{id}",
			Handler:     handler.Get,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
		{
			Method:      http.MethodPut,
			Path:        "/ingredients/{id}",
			Handler:     handler.Update,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
		{
			Method:      http.MethodGet,
			Path:        "/ingredients",
			Handler:     handler.List,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
		{
			Method:      http.MethodDelete,
			Path:        "/ingredients/{id}",
			Handler:     handler.Delete,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
		{
			Method:      http.MethodPut,
			Path:        "/ingredients/{id}/undelete",
			Handler:     handler.UnDelete,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
	}
	common.RegisterRoutes(router, routes)
}
