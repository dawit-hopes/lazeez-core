package folio

import (
	"encoding/json"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"
)

type FolioHandler interface {
	GetByBooking(w http.ResponseWriter, r *http.Request)
	Settle(w http.ResponseWriter, r *http.Request)
}

type folioHandler struct {
	service FolioService
	logger  config.Logger
}

func NewFolioHandler(service FolioService, logger config.Logger) FolioHandler {
	return &folioHandler{service: service, logger: logger}
}

func (h *folioHandler) GetByBooking(w http.ResponseWriter, r *http.Request) {
	bookingID := common.ParseID(r, "id")
	if bookingID == "" {
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}
	branchID, ok := middleware.GetBranchIDFromContext(r.Context())
	if !ok || branchID == "" {
		common.WriteErrorResponse(w, common.ErrUnAuthorized)
		return
	}

	bill, err := h.service.GetByBookingForBranch(r.Context(), bookingID, branchID)
	if err != nil {
		h.logger.Error("failed to get room bill", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       bill,
		Message:    "Room bill fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *folioHandler) Settle(w http.ResponseWriter, r *http.Request) {
	bookingID := common.ParseID(r, "id")
	if bookingID == "" {
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}
	branchID, ok := middleware.GetBranchIDFromContext(r.Context())
	if !ok || branchID == "" {
		common.WriteErrorResponse(w, common.ErrUnAuthorized)
		return
	}
	role, _ := middleware.GetRoleFromContext(r.Context())
	settledBy, _ := middleware.GetUserIDFromContext(r.Context())

	var req SettleBillInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode settle bill request", "error", err)
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}
	if err := req.Validate(); err != nil {
		h.logger.Error("failed to validate settle bill request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	bill, err := h.service.Settle(r.Context(), bookingID, branchID, role, settledBy, req.PaymentMethod)
	if err != nil {
		h.logger.Error("failed to settle room bill", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       bill,
		Message:    "Room bill settled successfully",
		StatusCode: http.StatusOK,
	})
}
