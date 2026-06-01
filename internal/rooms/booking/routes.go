package booking

import (
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewBookingRoutes(router chi.Router, handler BookingHandler, mw middleware.Middleware) {
	tokenMw := []func(next http.Handler) http.Handler{mw.ValidateToken}
	mutateMw := []func(next http.Handler) http.Handler{mw.ValidateToken, mw.RequireFrontDesk}

	routes := []common.Route{
		{
			Method:      http.MethodPost,
			Path:        "/bookings",
			Handler:     handler.Create,
			Middlewares: mutateMw,
		},
		{
			Method:      http.MethodGet,
			Path:        "/bookings",
			Handler:     handler.List,
			Middlewares: tokenMw,
		},
		{
			Method:      http.MethodGet,
			Path:        "/bookings/{id}",
			Handler:     handler.Get,
			Middlewares: tokenMw,
		},
		{
			Method:      http.MethodPut,
			Path:        "/bookings/{id}",
			Handler:     handler.Update,
			Middlewares: mutateMw,
		},
		{
			Method:      http.MethodPost,
			Path:        "/bookings/{id}/check-out",
			Handler:     handler.CheckOut,
			Middlewares: mutateMw,
		},
		{
			Method:      http.MethodPost,
			Path:        "/bookings/{id}/cancel",
			Handler:     handler.Cancel,
			Middlewares: mutateMw,
		},
		{
			Method:      http.MethodDelete,
			Path:        "/bookings/{id}",
			Handler:     handler.Delete,
			Middlewares: mutateMw,
		},
		{
			Method:      http.MethodPost,
			Path:        "/bookings/{id}/regenerate-passcode",
			Handler:     handler.ReGeneratePassCode,
			Middlewares: mutateMw,
		},
	}
	common.RegisterRoutes(router, routes)
}
