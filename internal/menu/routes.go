package menu

import (
	"lazeez-core/internal/common"

	"github.com/go-chi/chi/v5"
)

func NewMenuRoutes(router chi.Router, handler MenuHandler) {
	routes := []common.Route{
		{
			Method:  "POST",
			Path:    "/menus",
			Handler: handler.Create,
		},
		{
			Method:  "GET",
			Path:    "/menus/{id}",
			Handler: handler.Get,
		},
		{
			Method:  "PUT",
			Path:    "/menus/{id}",
			Handler: handler.Update,
		},
		{
			Method:  "DELETE",
			Path:    "/menus/{id}",
			Handler: handler.Delete,
		},
	}
	common.RegisterRoutes(router, routes)
}
