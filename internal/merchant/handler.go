package merchant

import (
	"errors"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"mime/multipart"
	"net/http"
)

type MerchantHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	GetAll(w http.ResponseWriter, r *http.Request)
}

type merchantHandler struct {
	merchantService MerchantService
	logger          config.Logger
}

func NewMerchantHandler(merchantService MerchantService, logger config.Logger) MerchantHandler {
	return &merchantHandler{merchantService: merchantService, logger: logger}
}

func (h *merchantHandler) parseMultipart(r *http.Request, limit int64) error {
	if err := r.ParseMultipartForm(limit); err != nil {
		h.logger.Error("Failed to parse multipart form", "error", err)
		return common.ErrInvalidMultipartForm
	}
	return nil
}

func (h *merchantHandler) parseRequest(r *http.Request, isRequired bool) (MerchantRequest, multipart.File, error) {
	var req MerchantRequest
	file, fileHeader, err := r.FormFile("logo")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			// If file is required, return a proper domain error instead of raw http error
			if isRequired {
				h.logger.Error("Logo file is required", "error", err)
				return req, nil, common.ErrMissingFile
			}
			// If not required, allow name-only updates
			req.Name = r.FormValue("name")
			return req, nil, nil
		}
		h.logger.Error("Failed to get logo file", "error", err)
		return req, nil, err
	}

	req.LogoHeader = *fileHeader
	req.Logo = file
	req.Name = r.FormValue("name")

	return req, file, nil
}

func (h *merchantHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := h.parseMultipart(r, 32<<20); err != nil {
		h.logger.Error("Failed to parse multipart form", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	req, file, err := h.parseRequest(r, true)
	if err != nil {
		h.logger.Error("Failed to parse request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	defer file.Close()

	if err := req.Validate(); err != nil {
		h.logger.Error("Failed to validate request body", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	if err := common.ValidateImage(req.LogoHeader); err != nil {
		h.logger.Error("Failed to validate image", "error", err)
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

func (h *merchantHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r, "id")
	merchant, err := h.merchantService.Get(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get merchant", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Data: merchant, Message: "Merchant fetched successfully", StatusCode: http.StatusOK})
}

func (h *merchantHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r, "id")
	if err := h.parseMultipart(r, 32<<20); err != nil {
		h.logger.Error("Failed to parse multipart form", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	req, file, err := h.parseRequest(r, false)
	if err != nil {
		h.logger.Error("Failed to parse request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	if file != nil {
		defer file.Close()
	}

	if req.Logo != nil {
		if err := common.ValidateImage(req.LogoHeader); err != nil {
			h.logger.Error("Failed to validate image", "error", err)
			common.WriteErrorResponse(w, err)
			return
		}
	}

	if IsEmpty(&req) {
		h.logger.Error("Name and logo are required")
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}

	err = h.merchantService.Update(r.Context(), id, req)
	if err != nil {
		h.logger.Error("Failed to update merchant", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Message: "Merchant updated successfully", StatusCode: http.StatusOK})
}

func (h *merchantHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r, "id")
	err := h.merchantService.Delete(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to delete merchant", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Message: "Merchant deleted successfully", StatusCode: http.StatusOK})
}

func (h *merchantHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	merchants, err := h.merchantService.GetAll(r.Context())
	if err != nil {
		h.logger.Error("Failed to get all merchants", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Data: merchants, Message: "Merchants fetched successfully", StatusCode: http.StatusOK})
}
