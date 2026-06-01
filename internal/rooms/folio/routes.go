package folio

import (
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewFolioRoutes(router chi.Router, handler FolioHandler, mw middleware.Middleware) {
	branchMw := []func(next http.Handler) http.Handler{mw.ValidateToken, mw.RequireBranch}

	routes := []common.Route{
		{
			Method:      http.MethodGet,
			Path:        "/branch/bookings/{id}/bill",
			Handler:     handler.GetByBooking,
			Middlewares: branchMw,
		},
		{
			Method:      http.MethodPost,
			Path:        "/branch/bookings/{id}/bill/settle",
			Handler:     handler.Settle,
			Middlewares: branchMw,
		},
	}
	common.RegisterRoutes(router, routes)
}
