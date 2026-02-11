package auth

import (
	"context"
	"encoding/json"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"net/http"
)

type AuthHandler interface {
	Login(w http.ResponseWriter, r *http.Request)
	FirstTimeLogin(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	ResetPassword(w http.ResponseWriter, r *http.Request)
}

type authHandler struct {
	authService AuthService
	logger      config.Logger
}

func NewAuthHandler(authService AuthService, logger config.Logger) AuthHandler {
	return &authHandler{authService: authService, logger: logger}
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

func (h *authHandler) FirstTimeLogin(w http.ResponseWriter, r *http.Request) {
	h.handleSetPassword(w, r, h.authService.FirstTimeLogin, "Password set successfully", "Failed to set password")
}

func (h *authHandler) Logout(w http.ResponseWriter, r *http.Request) {
	id := common.ParseID(r, "id")
	err := h.authService.Logout(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to logout", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Message: "Logout successful", StatusCode: http.StatusOK})
}

func (h *authHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	h.handleSetPassword(w, r, h.authService.ResetPassword, "Password reset successfully", "Failed to reset password")
}

func (h *authHandler) handleSetPassword(w http.ResponseWriter, r *http.Request, fn func(context.Context, SetPasswordRequest) (*LoginResponse, error), successMsg, errorMsg string) {
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
	loginResponse, err := fn(r.Context(), req)
	if err != nil {
		h.logger.Error(errorMsg, "error", err)
		common.WriteErrorResponse(w, err)
		return
	}
	common.WriteSuccessResponse(w, common.Response{Data: loginResponse, Message: successMsg, StatusCode: http.StatusOK})
}
