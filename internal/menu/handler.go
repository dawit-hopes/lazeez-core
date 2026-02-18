package menu

import (
	"encoding/json"
	"errors"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
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
}

type menuHandler struct {
	menuService MenuService
	logger      config.Logger
}

func NewMenuHandler(menuService MenuService, logger config.Logger) MenuHandler {
	return &menuHandler{
		menuService: menuService,
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
	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			if isRequired {
				h.logger.Error("Image file is required", "error", err)
				return req, nil, common.ErrMissingFile
			}
			// No image; still parse all other form fields for update
			h.parseFormFields(r, &req)
			branchID, ok := middleware.GetBranchIDFromContext(r.Context())
			if !ok {
				h.logger.Error("Failed to get branch ID from context", "error", common.ErrUnAuthorized)
				return req, nil, common.ErrUnAuthorized
			}
			req.BranchID = branchID
			return req, nil, nil
		}
		h.logger.Error("Failed to get image file", "error", err)
		return req, nil, err
	}

	req.ImageHeader = *fileHeader
	req.Image = file
	h.parseFormFields(r, &req)

	branchID, ok := middleware.GetBranchIDFromContext(r.Context())
	if !ok {
		h.logger.Error("Failed to get branch ID from context", "error", common.ErrUnAuthorized)
		return req, nil, common.ErrUnAuthorized
	}
	req.BranchID = branchID

	return req, file, nil
}

// parseFormFields fills MenuRequest from form values (name, description, price, ingredients, category_id, is_fasting, is_available).
func (h *menuHandler) parseFormFields(r *http.Request, req *MenuRequest) {
	req.Name = r.FormValue("name")
	req.Description = r.FormValue("description")
	if p := r.FormValue("price"); p != "" {
		if v, err := strconv.ParseFloat(p, 64); err == nil {
			req.Price = v
		}
	}
	if ing := r.FormValue("ingredients"); ing != "" {
		req.Ingredients = strings.Split(ing, ",")
	}
	req.CategoryID = r.FormValue("category_id")
	isFasting := r.FormValue("is_fasting") == "true"
	isAvailable := r.FormValue("is_available") == "true"
	req.IsFasting = &isFasting
	req.IsAvailable = &isAvailable

	// modifier_groups comes as a JSON string form field; decode into the request.
	if mg := r.FormValue("modifier_groups"); mg != "" {
		if err := json.Unmarshal([]byte(mg), &req.Modifiers); err != nil {
			h.logger.Error("Failed to parse modifier_groups", "error", err)
		}
	}
}

func (h *menuHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	if err := common.ValidateImage(req.ImageHeader); err != nil {
		h.logger.Error("Failed to validate image", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Error("Failed to validate create menu request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	createdMenu, err := h.menuService.Create(r.Context(), req)
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
	id := common.ParseID(r, "id")
	branchID, ok := middleware.GetBranchIDFromContext(r.Context())
	if !ok {
		h.logger.Error("Failed to get branch ID from context", "error", common.ErrUnAuthorized)
		common.WriteErrorResponse(w, common.ErrUnAuthorized)
		return
	}

	menu, err := h.menuService.Get(r.Context(), id, branchID)
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

	if req.IsEmpty() {
		h.logger.Error("No data to update", "error", common.ErrNoDataToUpdate)
		common.WriteErrorResponse(w, common.ErrNoDataToUpdate)
		return
	}

	if req.Image != nil {
		if err := common.ValidateImage(req.ImageHeader); err != nil {
			h.logger.Error("Failed to validate image", "error", err)
			common.WriteErrorResponse(w, err)
			return
		}
	}

	err = h.menuService.Update(r.Context(), id, req)
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
	id := common.ParseID(r, "id")
	branchID, ok := middleware.GetBranchIDFromContext(r.Context())
	if !ok {
		h.logger.Error("Failed to get branch ID from context", "error", common.ErrUnAuthorized)
		common.WriteErrorResponse(w, common.ErrUnAuthorized)
		return
	}

	if err := h.menuService.Delete(r.Context(), id, branchID); err != nil {
		h.logger.Error("Failed to delete menu", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Message:    "Menu deleted successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *menuHandler) UnDelete(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r, "id")
	branchID, ok := middleware.GetBranchIDFromContext(r.Context())
	if !ok {
		h.logger.Error("Failed to get branch ID from context", "error", common.ErrUnAuthorized)
		common.WriteErrorResponse(w, common.ErrUnAuthorized)
		return
	}
	err := h.menuService.UnDelete(r.Context(), id, branchID)
	if err != nil {
		h.logger.Error("Failed to undelete menu", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Message: "Menu undeleted successfully", StatusCode: http.StatusOK})
}

func (h *menuHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := common.ParseFilter(r)
	branchID, ok := middleware.GetBranchIDFromContext(r.Context())
	if !ok {
		h.logger.Error("Failed to get branch ID from context", "error", common.ErrUnAuthorized)
		common.WriteErrorResponse(w, common.ErrUnAuthorized)
		return
	}
	result, err := h.menuService.List(r.Context(), filter, branchID)
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
