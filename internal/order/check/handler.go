package check

import (
	"encoding/json"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"
)

type CheckHandler interface {
	ListOpen(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	Settle(w http.ResponseWriter, r *http.Request)
}

type checkHandler struct {
	service CheckService
	logger  config.Logger
}

func NewCheckHandler(service CheckService, logger config.Logger) CheckHandler {
	return &checkHandler{service: service, logger: logger}
}

func (h *checkHandler) ListOpen(w http.ResponseWriter, r *http.Request) {
	branchID, ok := middleware.GetBranchIDFromContext(r.Context())
	if !ok || branchID == "" {
		common.WriteErrorResponse(w, common.ErrUnAuthorized)
		return
	}
	checks, err := h.service.ListOpenForBranch(r.Context(), branchID)
	if err != nil {
		h.logger.Error("failed to list open checks", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       checks,
		Message:    "Checks fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *checkHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}
	branchID, ok := middleware.GetBranchIDFromContext(r.Context())
	if !ok || branchID == "" {
		common.WriteErrorResponse(w, common.ErrUnAuthorized)
		return
	}
	check, err := h.service.GetForBranch(r.Context(), id, branchID)
	if err != nil {
		h.logger.Error("failed to get check", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       check,
		Message:    "Check fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *checkHandler) Settle(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}
	branchID, ok := middleware.GetBranchIDFromContext(r.Context())
	if !ok || branchID == "" {
		common.WriteErrorResponse(w, common.ErrUnAuthorized)
		return
	}
	role, _ := middleware.GetRoleFromContext(r.Context())
	settledBy, _ := middleware.GetUserIDFromContext(r.Context())

	var req SettleCheckInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode settle check request", "error", err)
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}
	if err := req.Validate(); err != nil {
		h.logger.Error("failed to validate settle check request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	check, err := h.service.Settle(r.Context(), id, branchID, role, settledBy, req.PaymentMethod)
	if err != nil {
		h.logger.Error("failed to settle check", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       check,
		Message:    "Check settled successfully",
		StatusCode: http.StatusOK,
	})
}
