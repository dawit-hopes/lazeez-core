package auth

import (
	"context"
	"encoding/json"
	"lazeez-core/config"
	"lazeez-core/internal/common"
	"net/http"
	"time"
)

type AuthHandler interface {
	Login(w http.ResponseWriter, r *http.Request)
	FirstTimeLogin(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	ResetPassword(w http.ResponseWriter, r *http.Request)
	RefreshToken(w http.ResponseWriter, r *http.Request)
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

	h.setCookies(w, loginResponse.RefreshToken)
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

	h.removeCookies(w)
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

func (h *authHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		h.logger.Error("Failed to get refresh token from cookie", "error", err)
		common.WriteErrorResponse(w, common.ErrInvalidRequest)
		return
	}

	refreshToken := cookie.Value
	loginResponse, err := h.authService.RefreshToken(r.Context(), refreshToken)
	if err != nil {
		h.logger.Error("Failed to refresh token", "error", err)
		common.WriteErrorResponse(w, err)
		return
	}

	h.setCookies(w, loginResponse.RefreshToken)
	common.WriteSuccessResponse(w, common.Response{Data: loginResponse, Message: "Token refreshed successfully", StatusCode: http.StatusOK})
}

func (h *authHandler) setCookies(w http.ResponseWriter, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(time.Hour * 24 * 30),
	})
}

func (h *authHandler) removeCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(-time.Hour * 24 * 30),
	})
}
