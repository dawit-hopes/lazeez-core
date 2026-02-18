package category

import (
	"encoding/json"
	"lazeez-core/config"
	"lazeez-core/internal/common"
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

func (h *categoryHandler) Create(w http.ResponseWriter, r *http.Request) {

	var req CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	if err := req.Validate(true); err != nil {
		h.logger.Error("Failed to validate create category request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	err := h.categoryService.Create(r.Context(), req)
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
	var req CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	if err := req.Validate(false); err != nil {
		h.logger.Error("Failed to validate update category request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	err := h.categoryService.Update(r.Context(), id, req)
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
	filter := common.ParseFilter(r)
	result, err := h.categoryService.List(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list categories", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       result.Data,
		Meta:       &result.Meta,
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
