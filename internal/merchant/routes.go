package merchant

import (
	"lazeez-core/internal/common"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewMerchantRoutes(router chi.Router, handler MerchantHandler) {
	routes := []common.Route{
		{
			Method:  http.MethodPost,
			Path:    "/merchants",
			Handler: handler.Create,
		},
		{
			Method:  http.MethodGet,
			Path:    "/merchants/{id}",
			Handler: handler.Get,
		},
		{
			Method:  http.MethodPatch,
			Path:    "/merchants/{id}",
			Handler: handler.Update,
		},
		{
			Method:  http.MethodDelete,
			Path:    "/merchants/{id}",
			Handler: handler.Delete,
		},
		{
			Method:  http.MethodGet,
			Path:    "/merchants",
			Handler: handler.GetAll,
		},
	}
	common.RegisterRoutes(router, routes)
}
