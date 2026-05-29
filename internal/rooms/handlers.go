package rooms

import (
	"encoding/json"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"lazeez-core/internal/middleware"
	"net/http"
)

type RoomHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	Clone(w http.ResponseWriter, r *http.Request)
}

type roomHandler struct {
	roomService RoomService
	logger      config.Logger
}

func NewRoomHandler(roomService RoomService, logger config.Logger) RoomHandler {
	return &roomHandler{roomService: roomService, logger: logger}
}

func (h *roomHandler) contextValues(r *http.Request) (role, branchID, merchantID string) {
	role, _ = middleware.GetRoleFromContext(r.Context())
	branchID, _ = middleware.GetBranchIDFromContext(r.Context())
	merchantID, _ = middleware.GetMerchantIDFromContext(r.Context())
	return role, branchID, merchantID
}

func (h *roomHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req RoomRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	role, branchID, merchantID := h.contextValues(r)
	req.MerchantID = merchantID
	if err := req.Validate(false); err != nil {
		h.logger.Error("failed to validate request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	room, err := h.roomService.CreateRoom(r.Context(), req, role, branchID, merchantID)
	if err != nil {
		h.logger.Error("failed to create room", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       room,
		Message:    "Room created successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *roomHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r, "id")
	if id == "" {
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}

	role, branchID, merchantID := h.contextValues(r)
	room, err := h.roomService.GetRoom(r.Context(), id, role, branchID, merchantID)
	if err != nil {
		h.logger.Error("failed to get room", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       room,
		Message:    "Room fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *roomHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := common.ParseFilter(r)
	role, branchID, merchantID := h.contextValues(r)
	scope := ListScope(r.URL.Query().Get("scope"))

	result, err := h.roomService.ListRooms(r.Context(), filter, role, branchID, merchantID, scope)
	if err != nil {
		h.logger.Error("failed to list rooms", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       result.Data,
		Meta:       &result.Meta,
		Message:    "Rooms fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *roomHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r, "id")
	if id == "" {
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}

	var req RoomRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	if err := req.ValidateUpdate(); err != nil {
		h.logger.Error("failed to validate request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	role, branchID, merchantID := h.contextValues(r)
	err := h.roomService.UpdateRoom(r.Context(), id, req, role, branchID, merchantID)
	if err != nil {
		h.logger.Error("failed to update room", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Message:    "Room updated successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *roomHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r, "id")
	if id == "" {
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}

	role, branchID, merchantID := h.contextValues(r)
	err := h.roomService.DeleteRoom(r.Context(), id, role, branchID, merchantID)
	if err != nil {
		h.logger.Error("failed to delete room", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Message:    "Room deleted successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *roomHandler) Clone(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r, "id")
	if id == "" {
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}

	var req CloneRoomRequestDTO
	if r.Body != nil && r.ContentLength != 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.logger.Error("failed to decode clone request body", "error", err)
			common.WriteErrorResponse(w, err)
			return
		}
	}
	if err := req.Validate(); err != nil {
		h.logger.Error("failed to validate clone request", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	role, branchID, merchantID := h.contextValues(r)
	room, err := h.roomService.CloneRoom(r.Context(), id, req, role, branchID, merchantID)
	if err != nil {
		h.logger.Error("failed to clone room", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       room,
		Message:    "Room cloned successfully",
		StatusCode: http.StatusOK,
	})
}
