package order

import (
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewOrderRoutes(router chi.Router, handler OrderHandler, mw middleware.Middleware) {
	// Client routes: no JWT, sessionKey in header (X-Session-Key) or query (session_key)
	clientRoutes := []common.Route{
		{
			Method:  http.MethodPost,
			Path:    "/client/orders",
			Handler: handler.CreateClient,
			// No auth - sessionKey in body
		},
		{
			Method:      http.MethodGet,
			Path:        "/client/orders/{id}",
			Handler:     handler.GetClient,
			Middlewares: []func(next http.Handler) http.Handler{mw.RequireSessionKey},
		},
		{
			Method:      http.MethodGet,
			Path:        "/client/orders",
			Handler:     handler.ListClient,
			Middlewares: []func(next http.Handler) http.Handler{mw.RequireSessionKey},
		},
		{
			Method:      http.MethodPost,
			Path:        "/client/orders/{id}/cancel-payment",
			Handler:     handler.CancelPaymentClient,
			Middlewares: []func(next http.Handler) http.Handler{mw.RequireSessionKey},
		},
	}

	// Branch routes: JWT with branch_id (branch_manager)
	branchMw := []func(next http.Handler) http.Handler{mw.ValidateToken, mw.RequireBranch}
	branchRoutes := []common.Route{
		{
			Method:      http.MethodGet,
			Path:        "/branch/orders/{id}",
			Handler:     handler.GetBranch,
			Middlewares: branchMw,
		},
		{
			Method:      http.MethodGet,
			Path:        "/branch/orders",
			Handler:     handler.ListBranch,
			Middlewares: branchMw,
		},
		{
			Method:      http.MethodPut,
			Path:        "/branch/orders/{id}",
			Handler:     handler.UpdateBranch,
			Middlewares: branchMw,
		},
	}

	// Admin routes: JWT super_admin
	adminMw := []func(next http.Handler) http.Handler{mw.ValidateToken, mw.RequireSuperAdmin}
	adminRoutes := []common.Route{
		{
			Method:      http.MethodGet,
			Path:        "/admin/orders/{id}",
			Handler:     handler.GetAdmin,
			Middlewares: adminMw,
		},
		{
			Method:      http.MethodGet,
			Path:        "/admin/orders",
			Handler:     handler.ListAdmin,
			Middlewares: adminMw,
		},
		{
			Method:      http.MethodPut,
			Path:        "/admin/orders/{id}",
			Handler:     handler.UpdateAdmin,
			Middlewares: adminMw,
		},
	}

	webhookRoutes := []common.Route{
		{
			Method:      http.MethodPost,
			Path:        "/webhook/orders",
			Handler:     handler.ProcessPaymentWebHook,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateWebhook},
		},
	}

	common.RegisterRoutes(router, clientRoutes)
	common.RegisterRoutes(router, branchRoutes)
	common.RegisterRoutes(router, adminRoutes)
	common.RegisterRoutes(router, webhookRoutes)
}
