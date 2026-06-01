package table

import (
	"encoding/json"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"
)

type TableHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	RegenerateQRCode(w http.ResponseWriter, r *http.Request)
	AttachOrder(w http.ResponseWriter, r *http.Request)
	DetachOrder(w http.ResponseWriter, r *http.Request)
}

type tableHandler struct {
	tableService TableService
	logger       config.Logger
}

func NewTableHandler(tableService TableService, logger config.Logger) TableHandler {
	return &tableHandler{tableService: tableService, logger: logger}
}

func (h *tableHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req TableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Error("Failed to validate request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	// RequireBranch ensures branchID is set; user can only create tables for their branch
	branchID, _ := middleware.GetBranchIDFromContext(r.Context())
	if branchID != "" && req.BranchID != branchID {
		common.WriteErrorResponse(w, common.ErrUnAuthorized)
		return
	}

	tables, err := h.tableService.Create(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create tables", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       tables,
		Message:    "Tables created successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *tableHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		h.logger.Error("Table ID is required")
		common.WriteErrorResponse(w, err)
		return
	}
	branchID, _ := middleware.GetBranchIDFromContext(r.Context())

	table, err := h.tableService.Get(r.Context(), id, branchID)
	if err != nil {
		h.logger.Error("Failed to get table", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       table,
		Message:    "Table fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *tableHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		h.logger.Error("Table ID is required")
		common.WriteErrorResponse(w, err)
		return
	}
	branchID, _ := middleware.GetBranchIDFromContext(r.Context())

	err = h.tableService.Delete(r.Context(), id, branchID)
	if err != nil {
		h.logger.Error("Failed to delete table", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Message:    "Table deleted successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *tableHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := common.ParseFilter(r)
	branchID, _ := middleware.GetBranchIDFromContext(r.Context())

	result, err := h.tableService.List(r.Context(), filter, branchID)
	if err != nil {
		h.logger.Error("Failed to list tables", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       result.Data,
		Meta:       &result.Meta,
		Message:    "Tables fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *tableHandler) RegenerateQRCode(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		h.logger.Error("Table ID is required")
		common.WriteErrorResponse(w, err)
		return
	}
	branchID, _ := middleware.GetBranchIDFromContext(r.Context())

	err = h.tableService.RegenerateQRCode(r.Context(), id, branchID)
	if err != nil {
		h.logger.Error("Failed to regenerate QR code", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Message:    "QR code regenerated successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *tableHandler) AttachOrder(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		h.logger.Error("Table ID is required")
		common.WriteErrorResponse(w, err)
		return
	}
	branchID, _ := middleware.GetBranchIDFromContext(r.Context())

	var req AttachOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode attach order request", "error", err)
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}
	if req.OrderID == "" {
		h.logger.Error("Order ID is required")
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}

	err = h.tableService.AttachOrder(r.Context(), id, req.OrderID, branchID)
	if err != nil {
		h.logger.Error("Failed to attach order", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Message:    "Order attached successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *tableHandler) DetachOrder(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		h.logger.Error("Table ID is required")
		common.WriteErrorResponse(w, err)
		return
	}
	branchID, _ := middleware.GetBranchIDFromContext(r.Context())

	err = h.tableService.DetachOrder(r.Context(), id, branchID)
	if err != nil {
		h.logger.Error("Failed to detach order", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Message:    "Order detached successfully",
		StatusCode: http.StatusOK,
	})
}
