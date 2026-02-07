package auth

import (
	"lazeez-core/internal/common"

	"github.com/go-chi/chi/v5"
)

func NewAuthRoutes(router chi.Router, handler AuthHandler) {
	routes := []common.Route{
		{
			Method:  "POST",
			Path:    "/users",
			Handler: handler.CreateUser,
		},
		{
			Method:  "GET",
			Path:    "/users/{id}",
			Handler: handler.GetUserByID,
		},
		{
			Method:  "PUT",
			Path:    "/users/{id}",
			Handler: handler.UpdateUser,
		},
		{
			Method:  "DELETE",
			Path:    "/users/{id}",
			Handler: handler.DeleteUser,
		},
		{
			Method:  "POST",
			Path:    "/login",
			Handler: handler.Login,
		},
		{
			Method:  "POST",
			Path:    "/set-password",
			Handler: handler.SetPassword,
		},
		{
			Method:  "POST",
			Path:    "/user-look-up",
			Handler: handler.UserLookUp,
		},
	}
	common.RegisterRoutes(router, routes)
}
