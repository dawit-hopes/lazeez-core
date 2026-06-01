package order

import (
	"encoding/json"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"
)

type OrderHandler interface {
	// Client handlers (sessionKey auth)
	CreateClient(w http.ResponseWriter, r *http.Request)
	GetClient(w http.ResponseWriter, r *http.Request)
	ListClient(w http.ResponseWriter, r *http.Request)
	CancelPaymentClient(w http.ResponseWriter, r *http.Request)

	// Branch handlers (JWT with branch_id)
	GetBranch(w http.ResponseWriter, r *http.Request)
	ListBranch(w http.ResponseWriter, r *http.Request)
	ArchiveBranch(w http.ResponseWriter, r *http.Request)
	UpdateBranch(w http.ResponseWriter, r *http.Request)

	// Admin handlers (JWT super_admin)
	GetAdmin(w http.ResponseWriter, r *http.Request)
	ListAdmin(w http.ResponseWriter, r *http.Request)
	UpdateAdmin(w http.ResponseWriter, r *http.Request)

	// webhook handlers
	ProcessPaymentWebHook(w http.ResponseWriter, r *http.Request)
}

type orderHandler struct {
	orderService OrderService
	logger       config.Logger
}

func NewOrderHandler(orderService OrderService, logger config.Logger) OrderHandler {
	return &orderHandler{
		orderService: orderService,
		logger:       logger,
	}
}

// --- Client ---

func (h *orderHandler) CreateClient(w http.ResponseWriter, r *http.Request) {
	var req OrderInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Error("Failed to validate create order request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	order, err := h.orderService.CreateClient(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create order", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       order,
		Message:    "Order created successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *orderHandler) GetClient(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r, "id")
	sessionKey := SessionKeyFromRequest(r)
	if sessionKey == "" {
		common.WriteErrorResponse(w, common.ErrUnAuthorized)
		return
	}

	order, err := h.orderService.GetClient(r.Context(), id, sessionKey)
	if err != nil {
		h.logger.Error("Failed to get order", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       order,
		Message:    "Order fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *orderHandler) ListClient(w http.ResponseWriter, r *http.Request) {
	sessionKey := SessionKeyFromRequest(r)
	if sessionKey == "" {
		common.WriteErrorResponse(w, common.ErrUnAuthorized)
		return
	}

	filter := ParseOrderFilter(r)
	result, err := h.orderService.ListClient(r.Context(), filter, sessionKey)
	if err != nil {
		h.logger.Error("Failed to list orders", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       result.Data,
		Meta:       &result.Meta,
		Message:    "Orders fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *orderHandler) CancelPaymentClient(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r, "id")
	sessionKey := SessionKeyFromRequest(r)
	if sessionKey == "" {
		common.WriteErrorResponse(w, common.ErrUnAuthorized)
		return
	}

	if err := h.orderService.CancelPaymentClient(r.Context(), id, sessionKey); err != nil {
		h.logger.Error("Failed to cancel payment", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Message:    "Payment cancelled successfully",
		StatusCode: http.StatusOK,
	})
}

// --- Branch ---

func (h *orderHandler) GetBranch(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r, "id")
	branchID, ok := middleware.GetBranchIDFromContext(r.Context())
	if !ok || branchID == "" {
		common.WriteErrorResponse(w, common.ErrUnAuthorized)
		return
	}

	order, err := h.orderService.GetBranch(r.Context(), id, branchID)
	if err != nil {
		h.logger.Error("Failed to get order", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       order,
		Message:    "Order fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *orderHandler) ListBranch(w http.ResponseWriter, r *http.Request) {
	branchID, ok := middleware.GetBranchIDFromContext(r.Context())
	if !ok || branchID == "" {
		common.WriteErrorResponse(w, common.ErrUnAuthorized)
		return
	}

	filter := ParseOrderFilter(r)
	result, err := h.orderService.ListBranch(r.Context(), filter, branchID)
	if err != nil {
		h.logger.Error("Failed to list orders", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       result.Data,
		Meta:       &result.Meta,
		Message:    "Orders fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *orderHandler) ArchiveBranch(w http.ResponseWriter, r *http.Request) {
	branchID, ok := middleware.GetBranchIDFromContext(r.Context())
	if !ok || branchID == "" {
		common.WriteErrorResponse(w, common.ErrUnAuthorized)
		return
	}

	filter := ParseArchiveOrderFilter(r)
	result, err := h.orderService.ArchiveBranch(r.Context(), filter, branchID)
	if err != nil {
		h.logger.Error("Failed to list archived orders", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       result.Data,
		Meta:       &result.Meta,
		Message:    "Archived orders fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *orderHandler) UpdateBranch(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r, "id")
	branchID, ok := middleware.GetBranchIDFromContext(r.Context())
	if !ok || branchID == "" {
		common.WriteErrorResponse(w, common.ErrUnAuthorized)
		return
	}

	var req OrderUpdateInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Error("Failed to validate update order request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	if err := h.orderService.UpdateBranch(r.Context(), id, req, branchID); err != nil {
		h.logger.Error("Failed to update order", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Message:    "Order updated successfully",
		StatusCode: http.StatusOK,
	})
}

// --- Admin ---

func (h *orderHandler) GetAdmin(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r, "id")

	order, err := h.orderService.GetAdmin(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get order", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       order,
		Message:    "Order fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *orderHandler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	filter := ParseOrderFilter(r)
	if role, ok := middleware.GetRoleFromContext(r.Context()); ok && role == "super_branch_admin" {
		merchantID, ok := middleware.GetMerchantIDFromContext(r.Context())
		if !ok || merchantID == "" {
			common.WriteErrorResponse(w, common.ErrUnAuthorized)
			return
		}
		filter.MerchantID = merchantID
	}
	result, err := h.orderService.ListAdmin(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list orders", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       result.Data,
		Meta:       &result.Meta,
		Message:    "Orders fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *orderHandler) UpdateAdmin(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r, "id")

	var req OrderUpdateInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Error("Failed to validate update order request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	if err := h.orderService.UpdateAdmin(r.Context(), id, req); err != nil {
		h.logger.Error("Failed to update order", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Message:    "Order updated successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *orderHandler) ProcessPaymentWebHook(w http.ResponseWriter, r *http.Request) {
	rawBody, ok := r.Context().Value(middleware.WebhookBodyContextKey).([]byte)
	if !ok || len(rawBody) == 0 {
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}

	var req PaymentWebHookPayload
	if err := json.Unmarshal(rawBody, &req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}

	chapaSignature, _ := r.Context().Value(middleware.ChapaSignatureContextKey).(string)
	xSignature, _ := r.Context().Value(middleware.ChapaXSignatureContextKey).(string)
	if chapaSignature == "" && xSignature == "" {
		common.WriteErrorResponse(w, common.ErrUnAuthorized)
		return
	}

	if err := h.orderService.ProcessPaymentWebHook(r.Context(), req, rawBody, chapaSignature, xSignature); err != nil {
		h.logger.Error("Failed to process payment webhook", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Message:    "Payment webhook processed successfully",
		StatusCode: http.StatusOK,
	})
}
