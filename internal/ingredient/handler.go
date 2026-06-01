package ingredient

import (
	"encoding/json"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"net/http"
)

type IngredientHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	UnDelete(w http.ResponseWriter, r *http.Request)
}

type ingredientHandler struct {
	ingredientService IngredientService
	logger            config.Logger
}

func NewIngredientHandler(ingredientService IngredientService, logger config.Logger) IngredientHandler {
	return &ingredientHandler{
		ingredientService: ingredientService,
		logger:            logger,
	}
}

func (h *ingredientHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req IngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	if err := req.Validate(true, true); err != nil {
		h.logger.Error("Failed to validate request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	err := h.ingredientService.Create(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create ingredient", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Message:    "Ingredient created successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *ingredientHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}
	ingredientDTO, err := h.ingredientService.Get(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get ingredient", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       ingredientDTO,
		Message:    "Ingredient fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *ingredientHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}
	var req IngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	if req.IsEmpty() {
		h.logger.Error("Name and icon are required")
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}

	err = h.ingredientService.Update(r.Context(), id, req)
	if err != nil {
		h.logger.Error("Failed to update ingredient", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Message:    "Ingredient updated successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *ingredientHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := common.ParseFilter(r)
	result, err := h.ingredientService.List(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list ingredients", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       result.Data,
		Meta:       &result.Meta,
		Message:    "Ingredients fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *ingredientHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}
	err = h.ingredientService.Delete(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to delete ingredient", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Message:    "Ingredient deleted successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *ingredientHandler) UnDelete(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}
	err = h.ingredientService.UnDelete(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to undelete ingredient", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Message: "Ingredient undeleted successfully", StatusCode: http.StatusOK})
}
