package common

import "net/http"

var (
	BadRequest          = http.StatusBadRequest
	Unauthorized        = http.StatusUnauthorized
	NotFound            = http.StatusNotFound
	InternalServerError = http.StatusInternalServerError
	BadGateway          = http.StatusBadGateway
	ServiceUnavailable  = http.StatusServiceUnavailable
	GatewayTimeout      = http.StatusGatewayTimeout
	TooManyRequests     = http.StatusTooManyRequests
)

var (
	ErrUserAlreadyExists   = &Errors{Code: BadRequest, Message: "user already exists"}
	ErrPhoneNumberRequired = &Errors{Code: BadRequest, Message: "phone number is required"}
	ErrInvalidPhoneNumber  = &Errors{Code: BadRequest, Message: "invalid phone number"}
	ErrBranchNotFound      = &Errors{Code: NotFound, Message: "branch not found"}
	ErrUnAuthorized     = &Errors{Code: Unauthorized, Message: "unauthorized"}
	ErrPasswordRequired = &Errors{Code: BadRequest, Message: "password is required"}
	ErrInvalidPassword = &Errors{Code: BadRequest, Message: "invalid password"}
	ErrPasswordTooShort = &Errors{Code: BadRequest, Message: "password is too short"}
)
