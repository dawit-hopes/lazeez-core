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
	ErrUserAlreadyExists     = &Errors{Code: BadRequest, Message: "user already exists"}
	ErrPhoneNumberRequired   = &Errors{Code: BadRequest, Message: "phone number is required"}
	ErrInvalidPhoneNumber    = &Errors{Code: BadRequest, Message: "invalid phone number"}
	ErrBranchNotFound        = &Errors{Code: NotFound, Message: "branch not found"}
	ErrMenuNotFound          = &Errors{Code: NotFound, Message: "menu not found"}
	ErrUnAuthorized          = &Errors{Code: Unauthorized, Message: "unauthorized"}
	ErrPasswordRequired      = &Errors{Code: BadRequest, Message: "password is required"}
	ErrInvalidPassword       = &Errors{Code: BadRequest, Message: "invalid password"}
	ErrPasswordTooShort      = &Errors{Code: BadRequest, Message: "password is too short"}
	ErrMerchantAlreadyExists = &Errors{Code: BadRequest, Message: "merchant with this information already exists"}
	ErrMerchantNotFound      = &Errors{Code: NotFound, Message: "merchant not found"}
	ErrInternalServerError   = &Errors{Code: InternalServerError, Message: "internal server error"}
	ErrUserNotFound          = &Errors{Code: NotFound, Message: "user not found"}
	ErrImageSizeTooLarge     = &Errors{Code: BadRequest, Message: "image size is too large"}
	ErrInvalidImageType      = &Errors{Code: BadRequest, Message: "invalid image type"}
	ErrInvalidMultipartForm  = &Errors{Code: BadRequest, Message: "invalid multipart form"}
	ErrMissingFile           = &Errors{Code: BadRequest, Message: "missing file"}
	ErrNameAndLogoRequired   = &Errors{Code: BadRequest, Message: "name and logo are required"}
	ErrInvalidRequest        = &Errors{Code: BadRequest, Message: "invalid request"}
	ErrSecretKeyNotProvided  = &Errors{Code: BadRequest, Message: "secret key not provided"}
	ErrBranchAlreadyExists   = &Errors{Code: BadRequest, Message: "branch with this information already exists"}
	ErrUserWithInformationAlreadyExists = &Errors{Code: BadRequest, Message: "user with this information already exists"}
)
