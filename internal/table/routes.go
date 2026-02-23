package table

import (
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewTableRoutes(router chi.Router, handler TableHandler, mw middleware.Middleware) {
	// All table routes require JWT; branch-scoped routes use RequireBranch (super_admin can omit branch_id)
	branchMw := []func(next http.Handler) http.Handler{mw.ValidateToken, mw.RequireBranch}
	routes := []common.Route{
		{
			Method:      http.MethodPost,
			Path:        "/tables",
			Handler:     handler.Create,
			Middlewares: branchMw,
		},
		{
			Method:      http.MethodGet,
			Path:        "/tables/{id}",
			Handler:     handler.Get,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
		{
			Method:      http.MethodDelete,
			Path:        "/tables/{id}",
			Handler:     handler.Delete,
			Middlewares: branchMw,
		},
		{
			Method:      http.MethodGet,
			Path:        "/tables",
			Handler:     handler.List,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
		{
			Method:      http.MethodPost,
			Path:        "/tables/{id}/regenerate-qr-code",
			Handler:     handler.RegenerateQRCode,
			Middlewares: branchMw,
		},
		{
			Method:      http.MethodPost,
			Path:        "/tables/{id}/attach-order",
			Handler:     handler.AttachOrder,
			Middlewares: branchMw,
		},
		{
			Method:      http.MethodPost,
			Path:        "/tables/{id}/detach-order",
			Handler:     handler.DetachOrder,
			Middlewares: branchMw,
		},
	}
	common.RegisterRoutes(router, routes)
}
