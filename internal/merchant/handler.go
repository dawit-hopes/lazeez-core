package merchant

import (
	"encoding/json"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"net/http"
)

type MerchantHandler interface {
	Create(r *http.Request, w http.ResponseWriter)
	Get(r *http.Request, w http.ResponseWriter)
	Update(r *http.Request, w http.ResponseWriter)
	Delete(r *http.Request, w http.ResponseWriter)
	GetAll(r *http.Request, w http.ResponseWriter)
}

type merchantHandler struct {
	merchantService MerchantService
	logger          config.Logger
}

func NewMerchantHandler(merchantService MerchantService, logger config.Logger) MerchantHandler {
	return &merchantHandler{merchantService: merchantService, logger: logger}
}

func (h *merchantHandler) Create(r *http.Request, w http.ResponseWriter) {
	var req CreateMerchantRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Error("Failed to validate request body", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	merchant, err := h.merchantService.Create(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create merchant", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{Data: merchant, Message: "Merchant created successfully", StatusCode: http.StatusOK})

}

func (h *merchantHandler) Get(r *http.Request, w http.ResponseWriter) {
	id := common.ParseID(r)
	merchant, err := h.merchantService.Get(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get merchant", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Data: merchant, Message: "Merchant fetched successfully", StatusCode: http.StatusOK})
}

func (h *merchantHandler) Update(r *http.Request, w http.ResponseWriter) {
	var req UpdateMerchantRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Error("Failed to validate request body", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	merchant, err := h.merchantService.Update(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to update merchant", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Data: merchant, Message: "Merchant updated successfully", StatusCode: http.StatusOK})
}

func (h *merchantHandler) Delete(r *http.Request, w http.ResponseWriter) {
	id := common.ParseID(r)
	err := h.merchantService.Delete(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to delete merchant", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Message: "Merchant deleted successfully", StatusCode: http.StatusOK})
}

func (h *merchantHandler) GetAll(r *http.Request, w http.ResponseWriter) {
	merchants, err := h.merchantService.GetAll(r.Context())
	if err != nil {
		h.logger.Error("Failed to get all merchants", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Data: merchants, Message: "Merchants fetched successfully", StatusCode: http.StatusOK})
}
