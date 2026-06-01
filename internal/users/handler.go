package users

import (
	"encoding/json"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"net/http"
)

type UserHandler interface {
	CreateUser(w http.ResponseWriter, r *http.Request)
	CreateSuperAdminUser(w http.ResponseWriter, r *http.Request)
	GetUserByID(w http.ResponseWriter, r *http.Request)
	UpdateUser(w http.ResponseWriter, r *http.Request)
	DeleteUser(w http.ResponseWriter, r *http.Request)
	UserLookUp(w http.ResponseWriter, r *http.Request)
	GetAllUser(w http.ResponseWriter, r *http.Request)
	GetUserByBranchID(w http.ResponseWriter, r *http.Request)
	UnDeleteUser(w http.ResponseWriter, r *http.Request)
}

type userHandler struct {
	userService UserService
	logger      config.Logger
}

func NewUserHandler(userService UserService, logger config.Logger) UserHandler {
	return &userHandler{userService: userService, logger: logger}
}

func (h *userHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req UserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	if err := req.Validate(false); err != nil {
		h.logger.Error("Failed to validate request body", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	normalized, err := common.ValidatePhoneNumber(req.PhoneNumber)
	if err != nil {
		h.logger.Error("Invalid phone number", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	req.PhoneNumber = normalized
	err = h.userService.CreateUser(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create user", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Data: req, Message: "User created successfully", StatusCode: http.StatusOK})
}

func (h *userHandler) CreateSuperAdminUser(w http.ResponseWriter, r *http.Request) {
	var req SuperAdminUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	if err := req.Validate(false); err != nil {
		h.logger.Error("Failed to validate request body", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	normalized, err := common.ValidatePhoneNumber(req.PhoneNumber)
	if err != nil {
		h.logger.Error("Invalid phone number", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	req.PhoneNumber = normalized
	err = h.userService.CreateSuperAdminUser(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create super admin user", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Data: req, Message: "Super admin user created successfully", StatusCode: http.StatusOK})
}
func (h *userHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}
	user, err := h.userService.GetUserByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get user by ID", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Data: user, Message: "User fetched successfully", StatusCode: http.StatusOK})
}

func (h *userHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}
	var req UserRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	if err := req.Validate(true); err != nil {
		h.logger.Error("Failed to validate request body", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	if req.PhoneNumber != "" {
		normalized, err := common.ValidatePhoneNumber(req.PhoneNumber)
		if err != nil {
			h.logger.Error("Invalid phone number", "error", err)
			common.WriteErrorResponse(w, err)
			return
		}
		req.PhoneNumber = normalized
	}
	err = h.userService.UpdateUser(r.Context(), id, req)
	if err != nil {
		h.logger.Error("Failed to update user", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Message: "User updated successfully", StatusCode: http.StatusOK})
}

func (h *userHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}
	err = h.userService.DeleteUser(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to delete user", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Message: "User deleted successfully", StatusCode: http.StatusOK})
}

func (h *userHandler) UserLookUp(w http.ResponseWriter, r *http.Request) {
	var req UserLookUpRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	if err := req.Validate(); err != nil {
		h.logger.Error("Failed to validate request body", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	normalized, err := common.ValidatePhoneNumber(req.PhoneNumber)
	if err != nil {
		h.logger.Error("Invalid phone number", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	user, err := h.userService.UserLookUp(r.Context(), normalized)
	if err != nil {
		h.logger.Error("Failed to look up user by phone number", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Data: user, Message: "User looked up successfully", StatusCode: http.StatusOK})
}

func (h *userHandler) GetAllUser(w http.ResponseWriter, r *http.Request) {
	filter := common.ParseFilter(r)
	result, err := h.userService.GetAllUsers(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to get all users", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{
		Data:       result.Data,
		Meta:       &result.Meta,
		Message:    "Users fetched successfully",
		StatusCode: http.StatusOK,
	})
}

func (h *userHandler) GetUserByBranchID(w http.ResponseWriter, r *http.Request) {
	branchID, err := common.ParseID(r, "branchID")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}
	user, err := h.userService.GetUserByBranchID(r.Context(), branchID)
	if err != nil {
		h.logger.Error("Failed to get user by branch ID", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Data: user, Message: "User fetched successfully", StatusCode: http.StatusOK})
}

func (h *userHandler) UnDeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := common.ParseID(r, "id")
	if err != nil {
		common.WriteErrorResponse(w, err)
		return
	}
	err = h.userService.UnDeleteUser(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to undelete user", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Message: "User undeleted successfully", StatusCode: http.StatusOK})
}
