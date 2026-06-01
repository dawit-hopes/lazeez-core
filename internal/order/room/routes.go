package roomorder

import (
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRoomOrderRoutes(router chi.Router, handler RoomOrderHandler, mw middleware.Middleware) {
	// Client routes: no JWT; create uses JSON body, reads use reference + pass_code query params.
	clientRoutes := []common.Route{
		{
			Method:  http.MethodPost,
			Path:    "/client/room-orders",
			Handler: handler.CreateClient,
		},
		{
			Method:  http.MethodGet,
			Path:    "/client/room-orders/{id}",
			Handler: handler.GetClient,
		},
		{
			Method:  http.MethodGet,
			Path:    "/client/room-orders",
			Handler: handler.ListClient,
		},
	}

	// Branch routes: JWT with branch_id. Reads open to all branch roles;
	// updates restricted to front desk / room service in the service layer.
	branchMw := []func(next http.Handler) http.Handler{mw.ValidateToken, mw.RequireBranch}
	branchRoutes := []common.Route{
		{
			Method:      http.MethodGet,
			Path:        "/branch/room-orders/archive",
			Handler:     handler.ArchiveBranch,
			Middlewares: branchMw,
		},
		{
			Method:      http.MethodGet,
			Path:        "/branch/room-orders/{id}",
			Handler:     handler.GetBranch,
			Middlewares: branchMw,
		},
		{
			Method:      http.MethodGet,
			Path:        "/branch/room-orders",
			Handler:     handler.ListBranch,
			Middlewares: branchMw,
		},
		{
			Method:      http.MethodPut,
			Path:        "/branch/room-orders/{id}",
			Handler:     handler.UpdateBranch,
			Middlewares: branchMw,
		},
	}

	// Admin routes: super_admin (all) or super_branch_admin (merchant-scoped), read-only.
	adminRoutes := []common.Route{
		{
			Method:      http.MethodGet,
			Path:        "/admin/room-orders/{id}",
			Handler:     handler.GetAdmin,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken, mw.RequireOrderListAccess},
		},
		{
			Method:      http.MethodGet,
			Path:        "/admin/room-orders",
			Handler:     handler.ListAdmin,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken, mw.RequireOrderListAccess},
		},
	}

	common.RegisterRoutes(router, clientRoutes)
	common.RegisterRoutes(router, branchRoutes)
	common.RegisterRoutes(router, adminRoutes)
}
