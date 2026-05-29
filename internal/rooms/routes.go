package rooms

import (
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRoomRoutes(router chi.Router, handler RoomHandler, mw middleware.Middleware) {
	tokenMw := []func(next http.Handler) http.Handler{mw.ValidateToken}
	branchMw := []func(next http.Handler) http.Handler{mw.ValidateToken, mw.RequireBranch}

	routes := []common.Route{
		{
			Method:      http.MethodPost,
			Path:        "/rooms",
			Handler:     handler.Create,
			Middlewares: tokenMw,
		},
		{
			Method:      http.MethodGet,
			Path:        "/rooms/{id}",
			Handler:     handler.Get,
			Middlewares: tokenMw,
		},
		{
			Method:      http.MethodGet,
			Path:        "/rooms",
			Handler:     handler.List,
			Middlewares: tokenMw,
		},
		{
			Method:      http.MethodPut,
			Path:        "/rooms/{id}",
			Handler:     handler.Update,
			Middlewares: tokenMw,
		},
		{
			Method:      http.MethodDelete,
			Path:        "/rooms/{id}",
			Handler:     handler.Delete,
			Middlewares: tokenMw,
		},
		{
			Method:      http.MethodPost,
			Path:        "/rooms/{id}/clone",
			Handler:     handler.Clone,
			Middlewares: branchMw,
		},
	}
	common.RegisterRoutes(router, routes)
}
