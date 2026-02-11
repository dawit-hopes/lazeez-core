package auth

import (
	"encoding/json"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"net/http"
)

type AuthHandler interface {
	Login(w http.ResponseWriter, r *http.Request)
	SetPassword(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
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
