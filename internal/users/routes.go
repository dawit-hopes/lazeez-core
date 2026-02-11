package users

import (
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewUserRoutes(router chi.Router, handler UserHandler, mw middleware.Middleware) {
	routes := []common.Route{
		{
			Method:      http.MethodPost,
			Path:        "/users",
			Handler:     handler.CreateUser,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
		{
			Method:      http.MethodGet,
			Path:        "/users/{id}",
			Handler:     handler.GetUserByID,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
		{
			Method:      http.MethodPatch,
			Path:        "/users/{id}",
			Handler:     handler.UpdateUser,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
		{
			Method:      http.MethodDelete,
			Path:        "/users/{id}",
			Handler:     handler.DeleteUser,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
		{
			Method:  http.MethodPost,
			Path:    "/user-look-up",
			Handler: handler.UserLookUp,
		},
		{
			Method:      http.MethodGet,
			Path:        "/users",
			Handler:     handler.GetAllUser,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
		{
			Method:      http.MethodGet,
			Path:        "/users/branch/{branchID}",
			Handler:     handler.GetUserByBranchID,
			Middlewares: []func(next http.Handler) http.Handler{mw.ValidateToken},
		},
	}
	common.RegisterRoutes(router, routes)
}
