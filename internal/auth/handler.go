package auth

import (
	"encoding/json"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"net/http"
)

type AuthHandler interface {
	CreateUser(w http.ResponseWriter, r *http.Request)
	GetUserByID(w http.ResponseWriter, r *http.Request)
	UpdateUser(w http.ResponseWriter, r *http.Request)
	DeleteUser(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	SetPassword(w http.ResponseWriter, r *http.Request)
	UserLookUp(w http.ResponseWriter, r *http.Request)
}

type authHandler struct {
	authService AuthService
	logger      config.Logger
}

func NewAuthHandler(authService AuthService, logger config.Logger) AuthHandler {
	return &authHandler{authService: authService, logger: logger}
}

func (h *authHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req UserRequest
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
	err = h.authService.CreateUser(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create user", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Data: req, Message: "User created successfully", StatusCode: http.StatusOK})
}

func (h *authHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r)
	user, err := h.authService.GetUserByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get user by ID", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Data: user, Message: "User fetched successfully", StatusCode: http.StatusOK})
}

func (h *authHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var req UserRequest
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
	err = h.authService.UpdateUser(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to update user", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Data: req, Message: "User updated successfully", StatusCode: http.StatusOK})
}

func (h *authHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r)
	err := h.authService.DeleteUser(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to delete user", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Message: "User deleted successfully", StatusCode: http.StatusOK})
}

func (h *authHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
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
	loginResponse, err := h.authService.Login(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to login", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Data: loginResponse, Message: "Login successful", StatusCode: http.StatusOK})
}

func (h *authHandler) SetPassword(w http.ResponseWriter, r *http.Request) {
	var req SetPasswordRequest
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
	loginResponse, err := h.authService.SetPassword(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to set password", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Data: loginResponse, Message: "Password set successfully", StatusCode: http.StatusOK})
}

func (h *authHandler) UserLookUp(w http.ResponseWriter, r *http.Request) {
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
	user, err := h.authService.UserLookUp(r.Context(), req.PhoneNumber)
	if err != nil {
		h.logger.Error("Failed to look up user by phone number", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Data: user, Message: "User looked up successfully", StatusCode: http.StatusOK})
}
