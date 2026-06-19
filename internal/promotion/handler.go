package promotion

import (
	"errors"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"lazeez-core/internal/users"
	"mime/multipart"
	"net/http"
	"strings"
)

type PromotionHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	UnDelete(w http.ResponseWriter, r *http.Request)
}

type promotionHandler struct {
	promotionService PromotionService
	userService      users.UserService
	logger           config.Logger
}

func NewPromotionHandler(promotionService PromotionService, userService users.UserService, logger config.Logger) PromotionHandler {
	return &promotionHandler{
		promotionService: promotionService,
		userService:      userService,
		logger:           logger,
	}
}

func (h *promotionHandler) parseMultipart(r *http.Request, limit int64) error {
	if err := r.ParseMultipartForm(limit); err != nil {
		h.logger.Error("failed to parse multipart form", "error", err)
		return common.ErrInvalidMultipartForm
	}
	return nil
}

func (h *promotionHandler) parseRequest(r *http.Request, isRequired bool) (PromotionRequest, multipart.File, error) {
	var req PromotionRequest

	merchantID, err := h.resolveMerchantID(r)
	if err != nil {
		return req, nil, err
	}
	req.MerchantID = merchantID

	file, fileHeader, err := r.FormFile("banner_image")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			if isRequired {
				return req, nil, common.ErrMissingFile
			}
			if err := h.parseFormFields(r, &req); err != nil {
				return req, nil, err
			}
			return req, nil, nil
		}
		h.logger.Error("failed to get banner image file", "error", err)
		return req, nil, err
	}

	req.BannerImageHeader = *fileHeader
	req.BannerImage = file
	if err := h.parseFormFields(r, &req); err != nil {
		return req, nil, err
	}
	return req, file, nil
}

func (h *promotionHandler) parseFormFields(r *http.Request, req *PromotionRequest) error {
	req.Title = r.FormValue("title")
	req.Description = r.FormValue("description")

	if v := r.FormValue("is_active"); v != "" {
		isActive := v == "true"
		req.IsActive = &isActive
	}

	if start := r.FormValue("start_date"); start != "" {
		startDate, err := parseFormDate(start)
		if err != nil {
			h.logger.Error("failed to parse start_date", "error", err)
			return common.ErrInvalidRequest
		}
		req.StartDate = startDate
	}
	if end := r.FormValue("end_date"); end != "" {
		endDate, err := parseFormDate(end)
		if err != nil {
			h.logger.Error("failed to parse end_date", "error", err)
			return common.ErrInvalidRequest
		}
		req.EndDate = endDate
	}
	return nil
}

func (h *promotionHandler) resolveMerchantID(r *http.Request) (string, error) {
	if q := strings.TrimSpace(r.URL.Query().Get("merchant_id")); q != "" {
		return q, nil
	}
	if mid := strings.TrimSpace(r.FormValue("merchant_id")); mid != "" {
		return mid, nil
	}
	if mid, ok := middleware.GetMerchantIDFromContext(r.Context()); ok && mid != "" {
		return mid, nil
	}

	uid, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		return "", common.ErrUnAuthorized
	}

	userDTO, err := h.userService.GetUserByIDForLogin(r.Context(), uid)
	if err != nil {
		h.logger.Error("failed to resolve merchant for user", "user_id", uid, "error", err)
		return "", err
	}
	if userDTO.MerchantID == "" {
		return "", common.ErrBranchAdminMissingMerchant
	}
	return userDTO.MerchantID, nil
}

func (h *promotionHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := h.parseMultipart(r, 32<<20); err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	req, file, err := h.parseRequest(r, true)
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}
	defer file.Close()

	if err := req.Validate(true); err != nil {
		common.WriteErrorResponse(w, err)
		return
	}
	if err := common.ValidateImage(req.BannerImageHeader); err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	role, _ := middleware.GetRoleFromContext(r.Context())
	created, err := h.promotionService.Create(r.Context(), req, role)
	if err != nil {
		h.logger.Error("failed to create promotion", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       created,
		Message:    "Promotion created successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *promotionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	merchantID, err := h.resolveMerchantID(r)
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	role, _ := middleware.GetRoleFromContext(r.Context())
	promo, err := h.promotionService.Get(r.Context(), id, merchantID, role)
	if err != nil {
		h.logger.Error("failed to get promotion", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       promo,
		Message:    "Promotion fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *promotionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}
	if err := h.parseMultipart(r, 32<<20); err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	req, file, err := h.parseRequest(r, false)
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}
	if file != nil {
		defer file.Close()
	}
	if req.IsEmpty() {
		common.WriteErrorResponse(w, common.ErrNoDataToUpdate)
		return
	}
	if err := req.Validate(false); err != nil {
		common.WriteErrorResponse(w, err)
		return
	}
	if req.BannerImage != nil {
		if err := common.ValidateImage(req.BannerImageHeader); err != nil {
			common.WriteErrorResponse(w, err)
			return
		}
	}

	role, _ := middleware.GetRoleFromContext(r.Context())
	if err := h.promotionService.Update(r.Context(), id, req, role); err != nil {
		h.logger.Error("failed to update promotion", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Message:    "Promotion updated successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *promotionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	merchantID, err := h.resolveMerchantID(r)
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	role, _ := middleware.GetRoleFromContext(r.Context())
	if err := h.promotionService.Delete(r.Context(), id, merchantID, role); err != nil {
		h.logger.Error("failed to delete promotion", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Message:    "Promotion removed successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *promotionHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := common.ParseFilter(r)
	merchantID, err := h.resolveMerchantID(r)
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	role, _ := middleware.GetRoleFromContext(r.Context())
	result, err := h.promotionService.List(r.Context(), filter, merchantID, role)
	if err != nil {
		h.logger.Error("failed to list promotions", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       result.Data,
		Meta:       &result.Meta,
		Message:    "Promotions fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *promotionHandler) UnDelete(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	merchantID, err := h.resolveMerchantID(r)
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	role, _ := middleware.GetRoleFromContext(r.Context())
	if err := h.promotionService.UnDelete(r.Context(), id, merchantID, role); err != nil {
		h.logger.Error("failed to undelete promotion", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Message:    "Promotion restored successfully",
		StatusCode: http.StatusOK,
	})
}
