package category

import (
	"errors"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"mime/multipart"
	"net/http"
)

type CategoryHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	UnDelete(w http.ResponseWriter, r *http.Request)
}

type categoryHandler struct {
	categoryService CategoryService
	logger          config.Logger
}

func NewCategoryHandler(categoryService CategoryService, logger config.Logger) CategoryHandler {
	return &categoryHandler{
		categoryService: categoryService,
		logger:          logger,
	}
}

func (h *categoryHandler) parseMultipart(r *http.Request, limit int64) error {
	if err := r.ParseMultipartForm(limit); err != nil {
		h.logger.Error("Failed to parse multipart form", "error", err)
		return common.ErrInvalidMultipartForm
	}
	return nil
}

func (h *categoryHandler) parseRequest(r *http.Request, iconRequired bool) (CategoryRequest, multipart.File, error) {
	var req CategoryRequest
	file, fileHeader, err := r.FormFile("icon")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			if iconRequired {
				h.logger.Error("Icon file is required", "error", err)
				return req, nil, common.ErrMissingFile
			}
			req.Name = r.FormValue("name")
			return req, nil, nil
		}
		h.logger.Error("Failed to get icon file", "error", err)
		return req, nil, err
	}

	req.IconHeader = *fileHeader
	req.Icon = file
	req.Name = r.FormValue("name")

	return req, file, nil
}

func (h *categoryHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	if err := common.ValidateImage(req.IconHeader); err != nil {
		h.logger.Error("Failed to validate icon", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	if err := req.Validate(true); err != nil {
		h.logger.Error("Failed to validate create category request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	err = h.categoryService.Create(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create category", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Message:    "Category created successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *categoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r, "id")

	category, err := h.categoryService.Get(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get category", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       category,
		Message:    "Category fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *categoryHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	if req.Icon != nil {
		if err := common.ValidateImage(req.IconHeader); err != nil {
			h.logger.Error("Failed to validate icon", "error", err)
			common.WriteErrorResponse(w, err)
			return
		}
	}

	if err := req.Validate(false); err != nil {
		h.logger.Error("Failed to validate update category request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	err = h.categoryService.Update(r.Context(), id, req)
	if err != nil {
		h.logger.Error("Failed to update category", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Message:    "Category updated successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *categoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r, "id")

	if err := h.categoryService.Delete(r.Context(), id); err != nil {
		h.logger.Error("Failed to delete category", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Message:    "Category deleted successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *categoryHandler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categoryService.List(r.Context())
	if err != nil {
		h.logger.Error("Failed to list categories", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       categories,
		Message:    "Categories fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *categoryHandler) UnDelete(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r, "id")
	err := h.categoryService.UnDelete(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to undelete category", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Message: "Category undeleted successfully", StatusCode: http.StatusOK})
}
