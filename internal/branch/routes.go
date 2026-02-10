package branch

import (
	"lazeez-core/internal/common"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewBranchRoutes(router chi.Router, handler BranchHandler) {
	routes := []common.Route{
		{
			Method:  http.MethodPost,
			Path:    "/branches",
			Handler: handler.Create,
		},
		{
			Method:  http.MethodGet,
			Path:    "/branches/{id}",
			Handler: handler.Get,
		},
		{
			Method:  http.MethodPatch,
			Path:    "/branches/{id}",
			Handler: handler.Update,
		},
		{
			Method:  http.MethodDelete,
			Path:    "/branches/{id}",
			Handler: handler.Delete,
		},
		{
			Method:  http.MethodGet,
			Path:    "/branches",
			Handler: handler.GetAll,
		},
		{
			Method:  http.MethodGet,
			Path:    "/branches/merchant/{merchantID}",
			Handler: handler.GetAllByMerchantID,
		},
	}
	common.RegisterRoutes(router, routes)
}
