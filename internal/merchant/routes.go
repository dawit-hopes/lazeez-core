package merchant

import (
	"lazeez-core/internal/common"

	"github.com/go-chi/chi/v5"
)

func NewMerchantRoutes(router chi.Router, handler MerchantHandler) {
	routes := []common.Route{
		{
			Method:  "POST",
			Path:    "/merchants",
			Handler: handler.Create,
		},
		{
			Method:  "GET",
			Path:    "/merchants/{id}",
			Handler: handler.Get,
		},
		{
			Method:  "PUT",
			Path:    "/merchants/{id}",
			Handler: handler.Update,
		},
		{
			Method:  "DELETE",
			Path:    "/merchants/{id}",
			Handler: handler.Delete,
		},
		{
			Method:  "GET",
			Path:    "/merchants",
			Handler: handler.GetAll,
		},
	}
	common.RegisterRoutes(router, routes)
}
