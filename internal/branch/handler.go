package branch

import (
	"encoding/json"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"net/http"
)

type BranchHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	UnDelete(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	ListByMerchantID(w http.ResponseWriter, r *http.Request)
}

type branchHandler struct {
	branchService BranchService
	logger        config.Logger
}

func NewBranchHandler(branchService BranchService, logger config.Logger) BranchHandler {
	return &branchHandler{
		branchService: branchService,
		logger:        logger,
	}
}

func (h *branchHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateBranchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode create branch request body", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Error("Failed to validate create branch request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	normalized, err := common.ValidatePhoneNumber(req.PhoneNumber)
	if err != nil {
		h.logger.Error("Invalid phone number", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	req.PhoneNumber = normalized
	err = h.branchService.Create(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create branch", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Message:    "Branch created successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *branchHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r, "id")

	branchDTO, err := h.branchService.Get(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get branch", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       branchDTO,
		Message:    "Branch fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *branchHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r, "id")
	var req UpdateBranchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode update branch request body", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Error("Failed to validate update branch request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	if req.PhoneNumber != "" {
		normalized, err := common.ValidatePhoneNumber(req.PhoneNumber)
		if err != nil {
			h.logger.Error("Invalid phone number", "error", err)
			common.WriteErrorResponse(w, err)
			return
		}
		req.PhoneNumber = normalized
	}
	err := h.branchService.Update(r.Context(), id, req)
	if err != nil {
		h.logger.Error("Failed to update branch", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Message:    "Branch updated successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *branchHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r, "id")

	if err := h.branchService.Delete(r.Context(), id); err != nil {
		h.logger.Error("Failed to delete branch", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Message:    "Branch deleted successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *branchHandler) UnDelete(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r, "id")

	if err := h.branchService.UnDelete(r.Context(), id); err != nil {
		h.logger.Error("Failed to undelete branch", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Message:    "Branch undeleted successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *branchHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := common.ParseFilter(r)
	result, err := h.branchService.List(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list branches", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       result.Data,
		Meta:       &result.Meta,
		Message:    "Branches fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *branchHandler) ListByMerchantID(w http.ResponseWriter, r *http.Request) {
	merchantID := common.ParseID(r, "merchantID")
	filter := common.ParseFilter(r)
	h.logger.Info("Listing branches by merchant ID", "merchantID", merchantID)
	result, err := h.branchService.ListByMerchantID(r.Context(), merchantID, filter)
	if err != nil {
		h.logger.Error("Failed to list branches by merchant ID", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       result.Data,
		Meta:       &result.Meta,
		Message:    "Branches fetched successfully",
		StatusCode: http.StatusOK,
	})
}
