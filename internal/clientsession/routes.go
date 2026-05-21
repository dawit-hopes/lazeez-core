package clientsession

import (
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewClientSessionRoutes(router chi.Router, handler ClientSessionHandler, _ middleware.Middleware) {
	routes := []common.Route{
		{
			Method:  http.MethodPost,
			Path:    "/client/sessions",
			Handler: handler.Create,
		},
	}
	common.RegisterRoutes(router, routes)
}
