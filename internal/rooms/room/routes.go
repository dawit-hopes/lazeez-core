package room

import (
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRoomRoutes(router chi.Router, handler RoomHandler, mw middleware.Middleware) {
	tokenMw := []func(next http.Handler) http.Handler{mw.ValidateToken}
	manageMw := []func(next http.Handler) http.Handler{mw.ValidateToken, mw.RequireRoomManagement}

	routes := []common.Route{
		{
			Method:      http.MethodPost,
			Path:        "/rooms",
			Handler:     handler.Create,
			Middlewares: manageMw,
		},
		{
			Method:      http.MethodGet,
			Path:        "/rooms",
			Handler:     handler.List,
			Middlewares: tokenMw,
		},
		{
			Method:      http.MethodGet,
			Path:        "/rooms/{id}",
			Handler:     handler.Get,
			Middlewares: tokenMw,
		},
		{
			Method:      http.MethodPut,
			Path:        "/rooms/{id}",
			Handler:     handler.Update,
			Middlewares: manageMw,
		},
		{
			Method:      http.MethodDelete,
			Path:        "/rooms/{id}",
			Handler:     handler.Delete,
			Middlewares: manageMw,
		},
		{
			Method:      http.MethodPost,
			Path:        "/rooms/{id}/regenerate-qr-code",
			Handler:     handler.RegenerateQRCode,
			Middlewares: manageMw,
		},
	}
	common.RegisterRoutes(router, routes)
}
