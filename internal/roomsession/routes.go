package roomsession

import (
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRoomSessionRoutes(router chi.Router, handler RoomSessionHandler, _ middleware.Middleware) {
	routes := []common.Route{
		{
			Method:  http.MethodPost,
			Path:    "/client/room-sessions",
			Handler: handler.Create,
		},
	}
	common.RegisterRoutes(router, routes)
}
