package feedback

import (
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewFeedbackRoutes(router chi.Router, handler FeedbackHandler, mw middleware.Middleware) {
	sessionMw := []func(next http.Handler) http.Handler{mw.RequireSessionKey}
	staffMw := []func(next http.Handler) http.Handler{mw.ValidateToken, mw.RequireFeedbackListAccess}

	clientRoutes := []common.Route{
		{
			Method:      http.MethodPost,
			Path:        "/client/orders/{id}/rating",
			Handler:     handler.SubmitOrderRatingClient,
			Middlewares: sessionMw,
		},
		{
			Method:      http.MethodGet,
			Path:        "/client/orders/{id}/rating",
			Handler:     handler.GetOrderRatingClient,
			Middlewares: sessionMw,
		},
		{
			Method:      http.MethodPost,
			Path:        "/client/stay-rating",
			Handler:     handler.SubmitStayRatingClient,
			Middlewares: sessionMw,
		},
		{
			Method:      http.MethodGet,
			Path:        "/client/stay-rating",
			Handler:     handler.GetStayRatingClient,
			Middlewares: sessionMw,
		},
	}

	branchRoutes := []common.Route{
		{
			Method:      http.MethodGet,
			Path:        "/branch/order-ratings",
			Handler:     handler.ListOrderRatingsBranch,
			Middlewares: staffMw,
		},
		{
			Method:      http.MethodGet,
			Path:        "/branch/order-ratings/{id}",
			Handler:     handler.GetOrderRatingBranch,
			Middlewares: staffMw,
		},
		{
			Method:      http.MethodGet,
			Path:        "/branch/stay-ratings",
			Handler:     handler.ListStayRatingsBranch,
			Middlewares: staffMw,
		},
		{
			Method:      http.MethodGet,
			Path:        "/branch/stay-ratings/{id}",
			Handler:     handler.GetStayRatingBranch,
			Middlewares: staffMw,
		},
	}

	adminRoutes := []common.Route{
		{
			Method:      http.MethodGet,
			Path:        "/admin/order-ratings",
			Handler:     handler.ListOrderRatingsAdmin,
			Middlewares: staffMw,
		},
		{
			Method:      http.MethodGet,
			Path:        "/admin/order-ratings/{id}",
			Handler:     handler.GetOrderRatingAdmin,
			Middlewares: staffMw,
		},
		{
			Method:      http.MethodGet,
			Path:        "/admin/stay-ratings",
			Handler:     handler.ListStayRatingsAdmin,
			Middlewares: staffMw,
		},
		{
			Method:      http.MethodGet,
			Path:        "/admin/stay-ratings/{id}",
			Handler:     handler.GetStayRatingAdmin,
			Middlewares: staffMw,
		},
	}

	common.RegisterRoutes(router, clientRoutes)
	common.RegisterRoutes(router, branchRoutes)
	common.RegisterRoutes(router, adminRoutes)
}
