package clientsession

import (
	"encoding/json"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"net/http"
)

type ClientSessionHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
}

type clientSessionHandler struct {
	service ClientSessionService
	logger  config.Logger
}

func NewClientSessionHandler(service ClientSessionService, logger config.Logger) ClientSessionHandler {
	return &clientSessionHandler{service: service, logger: logger}
}

func (h *clientSessionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateSessionInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode create session request", "error", err)
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Error("failed to validate create session request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	session, err := h.service.Create(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to create client session", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       session,
		Message:    "Session created successfully",
		StatusCode: http.StatusOK,
	})
}
