package booking

import (
	"encoding/json"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"
)

type BookingHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	CheckOut(w http.ResponseWriter, r *http.Request)
	Cancel(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	ReGeneratePassCode(w http.ResponseWriter, r *http.Request)
}

type bookingHandler struct {
	bookingService BookingService
	logger         config.Logger
}

func NewBookingHandler(bookingService BookingService, logger config.Logger) BookingHandler {
	return &bookingHandler{bookingService: bookingService, logger: logger}
}

func (h *bookingHandler) contextValues(r *http.Request) (role, branchID, merchantID string) {
	role, _ = middleware.GetRoleFromContext(r.Context())
	branchID, _ = middleware.GetBranchIDFromContext(r.Context())
	merchantID, _ = middleware.GetMerchantIDFromContext(r.Context())
	return role, branchID, merchantID
}

func (h *bookingHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req BookingRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", "error", err)
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}
	if err := req.Validate(); err != nil {
		h.logger.Error("failed to validate request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	role, branchID, merchantID := h.contextValues(r)
	b, err := h.bookingService.CreateBooking(r.Context(), req, role, branchID, merchantID)
	if err != nil {
		h.logger.Error("failed to create booking", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       b,
		Message:    "Guest checked in successfully",
		StatusCode: http.StatusCreated,
	})
}

func (h *bookingHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	role, branchID, merchantID := h.contextValues(r)
	b, err := h.bookingService.GetBooking(r.Context(), id, role, branchID, merchantID)
	if err != nil {
		h.logger.Error("failed to get booking", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       b,
		Message:    "Booking fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *bookingHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := common.ParseFilter(r)
	role, branchID, merchantID := h.contextValues(r)

	result, err := h.bookingService.ListBookings(r.Context(), filter, role, branchID, merchantID)
	if err != nil {
		h.logger.Error("failed to list bookings", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       result.Data,
		Meta:       &result.Meta,
		Message:    "Bookings fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *bookingHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	var req BookingUpdateRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", "error", err)
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}
	if err := req.Validate(); err != nil {
		h.logger.Error("failed to validate request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	role, branchID, merchantID := h.contextValues(r)
	b, err := h.bookingService.UpdateBooking(r.Context(), id, req, role, branchID, merchantID)
	if err != nil {
		h.logger.Error("failed to update booking", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       b,
		Message:    "Booking updated successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *bookingHandler) CheckOut(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	role, branchID, merchantID := h.contextValues(r)
	b, err := h.bookingService.CheckOut(r.Context(), id, role, branchID, merchantID)
	if err != nil {
		h.logger.Error("failed to check out booking", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       b,
		Message:    "Guest checked out successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *bookingHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	role, branchID, merchantID := h.contextValues(r)
	b, err := h.bookingService.Cancel(r.Context(), id, role, branchID, merchantID)
	if err != nil {
		h.logger.Error("failed to cancel booking", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       b,
		Message:    "Booking cancelled successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *bookingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	role, branchID, merchantID := h.contextValues(r)
	if err := h.bookingService.DeleteBooking(r.Context(), id, role, branchID, merchantID); err != nil {
		h.logger.Error("failed to delete booking", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Message:    "Booking deleted successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *bookingHandler) ReGeneratePassCode(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}

	role, branchID, merchantID := h.contextValues(r)
	passcode, err := h.bookingService.ReGeneratePassCode(r.Context(), id, role, branchID, merchantID)
	if err != nil {
		h.logger.Error("failed to re-generate passcode", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       map[string]string{
			"passcode": *passcode,
		},
		Message:    "Passcode re-generated successfully",
		StatusCode: http.StatusOK,
	})
}
