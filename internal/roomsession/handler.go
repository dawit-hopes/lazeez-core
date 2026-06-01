package roomsession

import (
	"encoding/json"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"net/http"
)

type RoomSessionHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
}

type roomSessionHandler struct {
	service RoomSessionService
	logger  config.Logger
}

func NewRoomSessionHandler(service RoomSessionService, logger config.Logger) RoomSessionHandler {
	return &roomSessionHandler{service: service, logger: logger}
}

func (h *roomSessionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRoomSessionInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode create room session request", "error", err)
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Error("failed to validate create room session request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	session, err := h.service.Create(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to create room session", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       session,
		Message:    "Room session created successfully",
		StatusCode: http.StatusOK,
	})
}
