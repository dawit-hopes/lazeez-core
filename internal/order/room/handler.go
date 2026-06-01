package roomorder

import (
	"encoding/json"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"
)

type RoomOrderHandler interface {
	// Client (room reference + passcode)
	CreateClient(w http.ResponseWriter, r *http.Request)
	GetClient(w http.ResponseWriter, r *http.Request)
	ListClient(w http.ResponseWriter, r *http.Request)

	// Branch (JWT with branch_id)
	GetBranch(w http.ResponseWriter, r *http.Request)
	ListBranch(w http.ResponseWriter, r *http.Request)
	ArchiveBranch(w http.ResponseWriter, r *http.Request)
	UpdateBranch(w http.ResponseWriter, r *http.Request)

	// Admin
	GetAdmin(w http.ResponseWriter, r *http.Request)
	ListAdmin(w http.ResponseWriter, r *http.Request)
}

type roomOrderHandler struct {
	service RoomOrderService
	logger  config.Logger
}

func NewRoomOrderHandler(service RoomOrderService, logger config.Logger) RoomOrderHandler {
	return &roomOrderHandler{service: service, logger: logger}
}

// --- Client ---

func (h *roomOrderHandler) CreateClient(w http.ResponseWriter, r *http.Request) {
	var req RoomOrderInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode room order request", "error", err)
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}
	if err := req.Validate(); err != nil {
		h.logger.Error("failed to validate room order request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	order, err := h.service.CreateClient(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to create room order", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       order,
		Message:    "Room order created successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *roomOrderHandler) GetClient(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}
	reference, passCode, err := GuestAuthFromRequest(r)
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	order, err := h.service.GetClient(r.Context(), id, reference, passCode)
	if err != nil {
		h.logger.Error("failed to get room order", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       order,
		Message:    "Room order fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *roomOrderHandler) ListClient(w http.ResponseWriter, r *http.Request) {
	reference, passCode, err := GuestAuthFromRequest(r)
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	filter := ParseRoomOrderFilter(r)
	result, err := h.service.ListClient(r.Context(), filter, reference, passCode)
	if err != nil {
		h.logger.Error("failed to list room orders", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       result.Data,
		Meta:       &result.Meta,
		Message:    "Room orders fetched successfully",
		StatusCode: http.StatusOK,
	})
}

// --- Branch ---

func (h *roomOrderHandler) GetBranch(w http.ResponseWriter, r *http.Request) {
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

	order, err := h.service.GetBranch(r.Context(), id, branchID)
	if err != nil {
		h.logger.Error("failed to get room order", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       order,
		Message:    "Room order fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *roomOrderHandler) ListBranch(w http.ResponseWriter, r *http.Request) {
	branchID, ok := middleware.GetBranchIDFromContext(r.Context())
	if !ok || branchID == "" {
		common.WriteErrorResponse(w, common.ErrUnAuthorized)
		return
	}

	filter := ParseRoomOrderFilter(r)
	result, err := h.service.ListBranch(r.Context(), filter, branchID)
	if err != nil {
		h.logger.Error("failed to list room orders", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       result.Data,
		Meta:       &result.Meta,
		Message:    "Room orders fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *roomOrderHandler) ArchiveBranch(w http.ResponseWriter, r *http.Request) {
	branchID, ok := middleware.GetBranchIDFromContext(r.Context())
	if !ok || branchID == "" {
		common.WriteErrorResponse(w, common.ErrUnAuthorized)
		return
	}

	filter := ParseArchiveRoomOrderFilter(r)
	result, err := h.service.ArchiveBranch(r.Context(), filter, branchID)
	if err != nil {
		h.logger.Error("failed to list archived room orders", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       result.Data,
		Meta:       &result.Meta,
		Message:    "Archived room orders fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *roomOrderHandler) UpdateBranch(w http.ResponseWriter, r *http.Request) {
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

	var req RoomOrderUpdateInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode room order update request", "error", err)
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}
	if err := req.Validate(); err != nil {
		h.logger.Error("failed to validate room order update request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	if err := h.service.UpdateBranch(r.Context(), id, req, role, branchID); err != nil {
		h.logger.Error("failed to update room order", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Message:    "Room order updated successfully",
		StatusCode: http.StatusOK,
	})
}

// --- Admin ---

func (h *roomOrderHandler) GetAdmin(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}
	order, err := h.service.GetAdmin(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get room order", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       order,
		Message:    "Room order fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *roomOrderHandler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	filter := ParseRoomOrderFilter(r)
	if role, ok := middleware.GetRoleFromContext(r.Context()); ok && role == "super_branch_admin" {
		merchantID, ok := middleware.GetMerchantIDFromContext(r.Context())
		if !ok || merchantID == "" {
			common.WriteErrorResponse(w, common.ErrUnAuthorized)
			return
		}
		filter.MerchantID = merchantID
	}
	result, err := h.service.ListAdmin(r.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list room orders", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       result.Data,
		Meta:       &result.Meta,
		Message:    "Room orders fetched successfully",
		StatusCode: http.StatusOK,
	})
}
