package auth

import (
	"lazeez-core/internal/common"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewAuthRoutes(router chi.Router, handler AuthHandler) {
	routes := []common.Route{
		{
			Method:  http.MethodPost,
			Path:    "/users",
			Handler: handler.CreateUser,
		},
		{
			Method:  http.MethodGet,
			Path:    "/users/{id}",
			Handler: handler.GetUserByID,
		},
		{
			Method:  http.MethodPatch,
			Path:    "/users/{id}",
			Handler: handler.UpdateUser,
		},
		{
			Method:  http.MethodDelete,
			Path:    "/users/{id}",
			Handler: handler.DeleteUser,
		},
		{
			Method:  http.MethodPost,
			Path:    "/login",
			Handler: handler.Login,
		},
		{
			Method:  http.MethodPost,
			Path:    "/set-password",
			Handler: handler.SetPassword,
		},
		{
			Method:  http.MethodPost,
			Path:    "/user-look-up",
			Handler: handler.UserLookUp,
		},
		{
			Method:  http.MethodGet,
			Path:    "/users",
			Handler: handler.GetAllUser,
		},
		{
			Method:  http.MethodGet,
			Path:    "/users/branch/{branchID}",
			Handler: handler.GetUserByBranchID,
		},
	}
	common.RegisterRoutes(router, routes)
}
