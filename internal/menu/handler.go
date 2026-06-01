package menu

import (
	"encoding/json"
	"errors"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"lazeez-core/internal/users"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
)

type MenuHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	UnDelete(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	// public handlers
	ListMenus(w http.ResponseWriter, r *http.Request)
}

type menuHandler struct {
	menuService MenuService
	userService users.UserService
	logger      config.Logger
}

func NewMenuHandler(menuService MenuService, userService users.UserService, logger config.Logger) MenuHandler {
	return &menuHandler{
		menuService: menuService,
		userService: userService,
		logger:      logger,
	}
}

func (h *menuHandler) parseMultipart(r *http.Request, limit int64) error {
	if err := r.ParseMultipartForm(limit); err != nil {
		h.logger.Error("Failed to parse multipart form", "error", err)
		return common.ErrInvalidMultipartForm
	}
	return nil
}

func (h *menuHandler) parseRequest(r *http.Request, isRequired bool) (MenuRequest, multipart.File, error) {
	var req MenuRequest
	role, _ := middleware.GetRoleFromContext(r.Context())
	branchID, hasBranch := middleware.GetBranchIDFromContext(r.Context())

	merchantID, err := h.resolveMerchantID(r)
	if err != nil {
		return req, nil, err
	}

	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			if isRequired {
				h.logger.Error("Image file is required", "error", err)
				return req, nil, common.ErrMissingFile
			}
			if err := h.parseFormFields(r, &req); err != nil {
				return req, nil, err
			}
			if err := h.applyContextToRequest(&req, role, branchID, hasBranch, merchantID); err != nil {
				return req, nil, err
			}
			return req, nil, nil
		}
		h.logger.Error("Failed to get image file", "error", err)
		return req, nil, err
	}

	req.ImageHeader = *fileHeader
	req.Image = file
	if err := h.parseFormFields(r, &req); err != nil {
		return req, nil, err
	}

	if err := h.applyContextToRequest(&req, role, branchID, hasBranch, merchantID); err != nil {
		return req, nil, err
	}

	return req, file, nil
}

func (h *menuHandler) applyContextToRequest(req *MenuRequest, role, branchID string, hasBranch bool, merchantID string) error {
	if role == "super_branch_admin" && !hasBranch {
		if merchantID == "" {
			return common.ErrBranchAdminMissingMerchant
		}
		req.BranchID = ""
		req.MerchantID = merchantID
		return nil
	}
	if role == "super_admin" && !hasBranch {
		req.BranchID = ""
		if merchantID != "" {
			req.MerchantID = merchantID
		}
		return nil
	}
	if hasBranch {
		req.BranchID = branchID
		if merchantID != "" {
			req.MerchantID = merchantID
		}
		return nil
	}
	return common.ErrUnAuthorized
}

func (h *menuHandler) parseFormFields(r *http.Request, req *MenuRequest) error {
	req.Name = r.FormValue("name")
	req.Description = r.FormValue("description")
	if p := r.FormValue("price"); p != "" {
		if v, err := strconv.ParseFloat(p, 64); err == nil {
			req.Price = v
		}
	}
	if prep := r.FormValue("preparation_time"); prep != "" {
		if v, err := strconv.ParseFloat(prep, 64); err == nil {
			req.PreparationTime = v
		}
	}
	if ing := r.FormValue("ingredients"); ing != "" {
		req.Ingredients = strings.Split(ing, ",")
	}
	req.CategoryID = r.FormValue("category_id")
	if mid := strings.TrimSpace(r.FormValue("merchant_id")); mid != "" {
		req.MerchantID = mid
	}
	if v := r.FormValue("is_fasting"); v != "" {
		isFasting := v == "true"
		req.IsFasting = &isFasting
	}
	if v := r.FormValue("is_available"); v != "" {
		isAvailable := v == "true"
		req.IsAvailable = &isAvailable
	}

	if mg := r.FormValue("modifier_groups"); mg != "" {
		req.ModifiersSet = true
		if err := json.Unmarshal([]byte(mg), &req.Modifiers); err != nil {
			h.logger.Error("Failed to parse modifier_groups", "error", err)
			return common.ErrInvalidRequest
		}
	}
	return nil
}

func (h *menuHandler) parseScope(r *http.Request) ListScope {
	scope := ListScope(strings.TrimSpace(r.URL.Query().Get("scope")))
	switch scope {
	case ScopeMaster, ScopeBranchEffective, ScopeBranchManage, ScopeAllBranches:
		return scope
	default:
		return ""
	}
}

// resolveMerchantID returns merchant from query, form, JWT, or the authenticated user's profile.
func (h *menuHandler) resolveMerchantID(r *http.Request) (string, error) {
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
		h.logger.Error("Failed to resolve merchant for user", "user_id", uid, "error", err)
		return "", err
	}
	if userDTO.MerchantID == "" {
		h.logger.Error("User has no merchant_id", "user_id", uid, "role", userDTO.Role)
		return "", common.ErrBranchAdminMissingMerchant
	}
	return userDTO.MerchantID, nil
}

func (h *menuHandler) resolveBranchID(r *http.Request) (string, error) {
	if q := strings.TrimSpace(r.URL.Query().Get("branch_id")); q != "" {
		return q, nil
	}
	if branchID, ok := middleware.GetBranchIDFromContext(r.Context()); ok && branchID != "" {
		return branchID, nil
	}
	uid, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		return "", common.ErrUnAuthorized
	}
	userDTO, err := h.userService.GetUserByIDForLogin(r.Context(), uid)
	if err != nil {
		h.logger.Error("Failed to resolve branch for user", "user_id", uid, "error", err)
		return "", err
	}
	if userDTO.BranchID == "" {
		return "", common.ErrUnAuthorized
	}
	return userDTO.BranchID, nil
}

func (h *menuHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	if err := common.ValidateImage(req.ImageHeader); err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	if err := req.Validate(); err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	role, _ := middleware.GetRoleFromContext(r.Context())
	createdMenu, err := h.menuService.Create(r.Context(), req, role)
	if err != nil {
		h.logger.Error("Failed to create menu", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       createdMenu,
		Message:    "Menu created successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *menuHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}
	branchID, _ := middleware.GetBranchIDFromContext(r.Context())
	merchantID, err := h.resolveMerchantID(r)
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	menu, err := h.menuService.Get(r.Context(), id, branchID, merchantID)
	if err != nil {
		h.logger.Error("Failed to get menu", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       menu,
		Message:    "Menu fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *menuHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	if req.Image != nil {
		if err := common.ValidateImage(req.ImageHeader); err != nil {
			common.WriteErrorResponse(w, err)
			return
		}
	}

	role, _ := middleware.GetRoleFromContext(r.Context())
	err = h.menuService.Update(r.Context(), id, req, role)
	if err != nil {
		h.logger.Error("Failed to update menu", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Message:    "Menu updated successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *menuHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}
	branchID, _ := middleware.GetBranchIDFromContext(r.Context())
	merchantID, err := h.resolveMerchantID(r)
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}
	role, _ := middleware.GetRoleFromContext(r.Context())

	if err := h.menuService.Delete(r.Context(), id, branchID, merchantID, role); err != nil {
		h.logger.Error("Failed to delete menu", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Message:    "Menu removed successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *menuHandler) UnDelete(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}
	branchID, _ := middleware.GetBranchIDFromContext(r.Context())
	merchantID, err := h.resolveMerchantID(r)
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}
	role, _ := middleware.GetRoleFromContext(r.Context())

	err = h.menuService.UnDelete(r.Context(), id, branchID, merchantID, role)
	if err != nil {
		h.logger.Error("Failed to undelete menu", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Message: "Menu restored successfully", StatusCode: http.StatusOK})
}

func (h *menuHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := common.ParseFilter(r)
	role, _ := middleware.GetRoleFromContext(r.Context())
	scope := h.parseScope(r)

	branchID := ""
	if users.IsBranchStaffRoleString(role) {
		var err error
		branchID, err = h.resolveBranchID(r)
		if err != nil {
			common.WriteErrorResponse(w, err)
			return
		}
	} else {
		branchID, _ = middleware.GetBranchIDFromContext(r.Context())
	}

	merchantID, err := h.resolveMerchantID(r)
	if err != nil && (role == "super_branch_admin" || scope == ScopeMaster) {
		common.WriteErrorResponse(w, err)
		return
	}

	if qBranch := strings.TrimSpace(r.URL.Query().Get("branch_id")); qBranch != "" {
		branchID = qBranch
		if scope == ScopeMaster {
			scope = ScopeBranchEffective
		}
	}

	result, err := h.menuService.List(r.Context(), filter, branchID, merchantID, role, scope)
	if err != nil {
		h.logger.Error("Failed to list menus", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       result.Data,
		Meta:       &result.Meta,
		Message:    "Menus fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *menuHandler) ListMenus(w http.ResponseWriter, r *http.Request) {
	filter := common.ParseFilter(r)
	reference := r.URL.Query().Get("reference")

	if reference == "" {
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}

	result, err := h.menuService.ListMenus(r.Context(), filter, reference)
	if err != nil {
		h.logger.Error("Failed to list menus", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       result,
		Meta:       &result.Meta,
		Message:    "Menus fetched successfully",
		StatusCode: http.StatusOK,
	})
}
