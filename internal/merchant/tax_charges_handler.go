package merchant

import (
	"encoding/json"
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"
)

func (h *merchantHandler) GetTaxCharges(w http.ResponseWriter, r *http.Request) {
	merchantID, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	role, _ := middleware.GetRoleFromContext(r.Context())
	callerMerchantID, _ := middleware.GetMerchantIDFromContext(r.Context())

	taxCharges, err := h.merchantService.GetTaxCharges(r.Context(), merchantID, role, callerMerchantID)
	if err != nil {
		h.logger.Error("Failed to get merchant tax charges", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       taxCharges,
		Message:    "Tax charges fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *merchantHandler) UpdateTaxCharges(w http.ResponseWriter, r *http.Request) {
	merchantID, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	var req TaxChargesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode tax charges request", "error", err)
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}

	role, _ := middleware.GetRoleFromContext(r.Context())
	callerMerchantID, _ := middleware.GetMerchantIDFromContext(r.Context())

	taxCharges, err := h.merchantService.UpdateTaxCharges(r.Context(), merchantID, req, role, callerMerchantID)
	if err != nil {
		h.logger.Error("Failed to update merchant tax charges", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       taxCharges,
		Message:    "Tax charges updated successfully",
		StatusCode: http.StatusOK,
	})
}
