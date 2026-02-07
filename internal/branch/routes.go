package branch

import (
	"lazeez-core/internal/common"

	"github.com/go-chi/chi/v5"
)

func NewBranchRoutes(router chi.Router, handler BranchHandler) {
	routes := []common.Route{
		{
			Method:  "POST",
			Path:    "/branches",
			Handler: handler.Create,
		},
		{
			Method:  "GET",
			Path:    "/branches/{id}",
			Handler: handler.Get,
		},
		{
			Method:  "PUT",
			Path:    "/branches/{id}",
			Handler: handler.Update,
		},
		{
			Method:  "DELETE",
			Path:    "/branches/{id}",
			Handler: handler.Delete,
		},
	}
	common.RegisterRoutes(router, routes)
}
