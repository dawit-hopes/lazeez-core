package feedback

import (
	"encoding/json"
	"net/http"

	"lazeez-core/config"
	"lazeez-core/internal/common"
	order "lazeez-core/internal/order/table"
)

type FeedbackHandler interface {
	SubmitOrderRatingClient(w http.ResponseWriter, r *http.Request)
	GetOrderRatingClient(w http.ResponseWriter, r *http.Request)
	SubmitStayRatingClient(w http.ResponseWriter, r *http.Request)
	GetStayRatingClient(w http.ResponseWriter, r *http.Request)

	GetOrderRatingBranch(w http.ResponseWriter, r *http.Request)
	ListOrderRatingsBranch(w http.ResponseWriter, r *http.Request)
	GetStayRatingBranch(w http.ResponseWriter, r *http.Request)
	ListStayRatingsBranch(w http.ResponseWriter, r *http.Request)

	GetOrderRatingAdmin(w http.ResponseWriter, r *http.Request)
	ListOrderRatingsAdmin(w http.ResponseWriter, r *http.Request)
	GetStayRatingAdmin(w http.ResponseWriter, r *http.Request)
	ListStayRatingsAdmin(w http.ResponseWriter, r *http.Request)
}

type feedbackHandler struct {
	feedbackService FeedbackService
	logger          config.Logger
}

func NewFeedbackHandler(feedbackService FeedbackService, logger config.Logger) FeedbackHandler {
	return &feedbackHandler{
		feedbackService: feedbackService,
		logger:          logger,
	}
}

func (h *feedbackHandler) SubmitOrderRatingClient(w http.ResponseWriter, r *http.Request) {
	orderID, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	var req RatingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode order rating request", "error", err)
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}
	if err := req.Validate(); err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	sessionKey := order.SessionKeyFromRequest(r)
	rating, err := h.feedbackService.SubmitOrderRating(r.Context(), orderID, sessionKey, req)
	if err != nil {
		h.logger.Error("failed to submit order rating", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       rating,
		Message:    "Order rating submitted successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *feedbackHandler) GetOrderRatingClient(w http.ResponseWriter, r *http.Request) {
	orderID, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	sessionKey := order.SessionKeyFromRequest(r)
	rating, err := h.feedbackService.GetOrderRatingClient(r.Context(), orderID, sessionKey)
	if err != nil {
		h.logger.Error("failed to get order rating", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	message := "Order rating fetched successfully"
	if rating == nil {
		message = "Order rating not found"
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       rating,
		Message:    message,
		StatusCode: http.StatusOK,
	})
}

func (h *feedbackHandler) SubmitStayRatingClient(w http.ResponseWriter, r *http.Request) {
	var req StayRatingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode stay rating request", "error", err)
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}
	if err := req.Validate(); err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	sessionKey := order.SessionKeyFromRequest(r)
	rating, err := h.feedbackService.SubmitStayRating(r.Context(), sessionKey, req)
	if err != nil {
		h.logger.Error("failed to submit stay rating", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       rating,
		Message:    "Stay rating submitted successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *feedbackHandler) GetStayRatingClient(w http.ResponseWriter, r *http.Request) {
	sessionKey := order.SessionKeyFromRequest(r)
	rating, err := h.feedbackService.GetStayRatingClient(r.Context(), sessionKey)
	if err != nil {
		h.logger.Error("failed to get stay rating", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	message := "Stay rating fetched successfully"
	if rating == nil {
		message = "Stay rating not found"
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       rating,
		Message:    message,
		StatusCode: http.StatusOK,
	})
}

func (h *feedbackHandler) GetOrderRatingBranch(w http.ResponseWriter, r *http.Request) {
	h.getOrderRatingStaff(w, r, "branch")
}

func (h *feedbackHandler) ListOrderRatingsBranch(w http.ResponseWriter, r *http.Request) {
	h.listOrderRatingsStaff(w, r, "branch")
}

func (h *feedbackHandler) GetStayRatingBranch(w http.ResponseWriter, r *http.Request) {
	h.getStayRatingStaff(w, r, "branch")
}

func (h *feedbackHandler) ListStayRatingsBranch(w http.ResponseWriter, r *http.Request) {
	h.listStayRatingsStaff(w, r, "branch")
}

func (h *feedbackHandler) GetOrderRatingAdmin(w http.ResponseWriter, r *http.Request) {
	h.getOrderRatingStaff(w, r, "admin")
}

func (h *feedbackHandler) ListOrderRatingsAdmin(w http.ResponseWriter, r *http.Request) {
	h.listOrderRatingsStaff(w, r, "admin")
}

func (h *feedbackHandler) GetStayRatingAdmin(w http.ResponseWriter, r *http.Request) {
	h.getStayRatingStaff(w, r, "admin")
}

func (h *feedbackHandler) ListStayRatingsAdmin(w http.ResponseWriter, r *http.Request) {
	h.listStayRatingsStaff(w, r, "admin")
}

func (h *feedbackHandler) getOrderRatingStaff(w http.ResponseWriter, r *http.Request, source string) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	filter := ParseRatingFilter(r)
	scope, err := ResolveListScope(r.Context(), filter.BranchID)
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	rating, err := h.feedbackService.GetOrderRatingStaff(r.Context(), id, scope)
	if err != nil {
		h.logger.Error("failed to get order rating for staff", "source", source, "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       rating,
		Message:    "Order rating fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *feedbackHandler) getStayRatingStaff(w http.ResponseWriter, r *http.Request, source string) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	filter := ParseRatingFilter(r)
	scope, err := ResolveListScope(r.Context(), filter.BranchID)
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	rating, err := h.feedbackService.GetStayRatingStaff(r.Context(), id, scope)
	if err != nil {
		h.logger.Error("failed to get stay rating for staff", "source", source, "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	common.WriteSuccessResponse(w, common.Response{
		Data:       rating,
		Message:    "Stay rating fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *feedbackHandler) listOrderRatingsStaff(w http.ResponseWriter, r *http.Request, source string) {
	filter := ParseRatingFilter(r)
	if err := ValidateRatingDateRange(filter); err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	scope, err := ResolveListScope(r.Context(), filter.BranchID)
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	result, err := h.feedbackService.ListOrderRatingsStaff(r.Context(), filter, scope)
	if err != nil {
		h.logger.Error("failed to list order ratings for staff", "source", source, "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	writeRatingListResponse(w, result.Data, result.Meta, "Order ratings fetched successfully")
}

func (h *feedbackHandler) listStayRatingsStaff(w http.ResponseWriter, r *http.Request, source string) {
	filter := ParseRatingFilter(r)
	if err := ValidateRatingDateRange(filter); err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	scope, err := ResolveListScope(r.Context(), filter.BranchID)
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	result, err := h.feedbackService.ListStayRatingsStaff(r.Context(), filter, scope)
	if err != nil {
		h.logger.Error("failed to list stay ratings for staff", "source", source, "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	writeRatingListResponse(w, result.Data, result.Meta, "Stay ratings fetched successfully")
}

func writeRatingListResponse[T any](w http.ResponseWriter, data []T, meta common.PaginationMeta, message string) {
	if data == nil {
		data = []T{}
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       data,
		Meta:       &meta,
		Message:    message,
		StatusCode: http.StatusOK,
	})
}
