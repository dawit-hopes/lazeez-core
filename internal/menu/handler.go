package menu

import (
	"errors"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"mime/multipart"
	"net/http"
)

type MenuHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	UnDelete(w http.ResponseWriter, r *http.Request)
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
			// If file is required, return a proper domain error instead of raw http error
			if isRequired {
				h.logger.Error("Image file is required", "error", err)
				return req, nil, common.ErrMissingFile
			}
			// If not required, allow name-only updates
			req.Name = r.FormValue("name")
			return req, nil, nil
		}
		h.logger.Error("Failed to get image file", "error", err)
		return req, nil, err
	}

	req.ImageHeader = *fileHeader
	req.Image = file
	req.Name = r.FormValue("name")

	return req, file, nil
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

	menu, err := h.menuService.Create(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create menu", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       menu,
		Message:    "Menu created successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *menuHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r, "id")

	menu, err := h.menuService.Get(r.Context(), id)
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

	if req.Image != nil {
		if err := common.ValidateImage(req.ImageHeader); err != nil {
			h.logger.Error("Failed to validate image", "error", err)
			common.WriteErrorResponse(w, err)
			return
		}
	}

	menu, err := h.menuService.Update(r.Context(), id, req)
	if err != nil {
		h.logger.Error("Failed to update menu", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       menu,
		Message:    "Menu updated successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *menuHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r, "id")

	if err := h.menuService.Delete(r.Context(), id); err != nil {
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
	err := h.menuService.UnDelete(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to undelete menu", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Message: "Menu undeleted successfully", StatusCode: http.StatusOK})
}