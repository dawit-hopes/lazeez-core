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
	ErrUserAlreadyExists                = &Errors{Code: BadRequest, Message: "user already exists"}
	ErrPhoneNumberRequired              = &Errors{Code: BadRequest, Message: "phone number is required"}
	ErrInvalidPhoneNumber               = &Errors{Code: BadRequest, Message: "invalid phone number"}
	ErrBranchNotFound                   = &Errors{Code: NotFound, Message: "branch not found"}
	ErrMenuNotFound                     = &Errors{Code: NotFound, Message: "menu not found"}
	ErrCategoryNotFound                 = &Errors{Code: NotFound, Message: "category not found"}
	ErrIngredientNotFound               = &Errors{Code: NotFound, Message: "ingredient not found"}
	ErrModifierGroupNotFound            = &Errors{Code: NotFound, Message: "modifier group not found"}
	ErrModifierOptionNotFound           = &Errors{Code: NotFound, Message: "modifier option not found"}
	ErrUnAuthorized                     = &Errors{Code: Unauthorized, Message: "unauthorized"}
	ErrPasswordRequired                 = &Errors{Code: BadRequest, Message: "password is required and cannot be empty"}
	ErrInvalidPassword                  = &Errors{Code: BadRequest, Message: "password is invalid"}
	ErrPasswordTooShort                 = &Errors{Code: BadRequest, Message: "password is too short; it must be at least 9 characters long"}
	ErrPasswordMissingLetter            = &Errors{Code: BadRequest, Message: "password must contain at least one letter (a-z or A-Z)"}
	ErrPasswordMissingNumberOrSpecial   = &Errors{Code: BadRequest, Message: "password must contain at least one number (0-9) or one special character"}
	ErrMerchantAlreadyExists            = &Errors{Code: BadRequest, Message: "merchant with this information already exists"}
	ErrMerchantNotFound                 = &Errors{Code: NotFound, Message: "merchant not found"}
	ErrInternalServerError              = &Errors{Code: InternalServerError, Message: "internal server error"}
	ErrUserNotFound                     = &Errors{Code: NotFound, Message: "user not found"}
	ErrImageSizeTooLarge                = &Errors{Code: BadRequest, Message: "image size is too large"}
	ErrInvalidImageType                 = &Errors{Code: BadRequest, Message: "invalid image type"}
	ErrInvalidMultipartForm             = &Errors{Code: BadRequest, Message: "invalid multipart form"}
	ErrMissingFile                      = &Errors{Code: BadRequest, Message: "missing file"}
	ErrNameAndLogoRequired              = &Errors{Code: BadRequest, Message: "name and logo are required"}
	ErrInvalidRequest                   = &Errors{Code: BadRequest, Message: "invalid request"}
	ErrSecretKeyNotProvided             = &Errors{Code: BadRequest, Message: "secret key not provided"}
	ErrBranchAlreadyExists              = &Errors{Code: BadRequest, Message: "branch with this information already exists"}
	ErrUserWithInformationAlreadyExists = &Errors{Code: BadRequest, Message: "user with this information already exists"}
	ErrSessionAlreadyExists             = &Errors{Code: BadRequest, Message: "session already exists"}
	ErrSessionNotFound                  = &Errors{Code: NotFound, Message: "session not found"}
	ErrUserLocked                       = &Errors{Code: Unauthorized, Message: "user is locked"}
	ErrUserNotFirstTimeLogin            = &Errors{Code: BadRequest, Message: "user is not first time login."}
	ErrUserIsFirstTimeLogin             = &Errors{Code: BadRequest, Message: "user is first time user."}
	ErrSessionRevoked                   = &Errors{Code: Unauthorized, Message: "session is revoked"}
	ErrInvalidRefreshToken              = &Errors{Code: Unauthorized, Message: "invalid refresh token"}
	ErrWrongUsernameOrPassword          = &Errors{Code: Unauthorized, Message: "wrong username or password"}
	ErrRefreshTokenNotFound             = &Errors{Code: BadRequest, Message: "refresh token not found"}
	ErrCategoryAlreadyExists            = &Errors{Code: BadRequest, Message: "category with this information already exists"}
	ErrIngredientAlreadyExists          = &Errors{Code: BadRequest, Message: "ingredient with this information already exists"}
	ErrNoDataToUpdate                   = &Errors{Code: BadRequest, Message: "no data to update"}
	ErrMenuAlreadyExists                = &Errors{Code: BadRequest, Message: "menu with this information already exists"}
	ErrMerchentAlreadyDeleted           = &Errors{Code: BadRequest, Message: "merchant is already deleted"}
	ErrOrderItemNotFound                = &Errors{Code: NotFound, Message: "order item not found"}
	ErrOrderNotFound                    = &Errors{Code: NotFound, Message: "order not found"}
	ErrOrderAlreadyExists               = &Errors{Code: BadRequest, Message: "order with this information already exists"}
	ErrOrderItemAlreadyExists           = &Errors{Code: BadRequest, Message: "order item with this information already exists"}
	ErrMenuPriceMismatch                = &Errors{Code: BadRequest, Message: "menu item price does not match order item price"}
	ErrMenuItemsUnavailable             = &Errors{Code: BadRequest, Message: "some items in your order are no longer available"}
	ErrOrderTotalMismatch               = &Errors{Code: BadRequest, Message: "order total does not match sum of order items"}
	ErrOrderNumberExhausted             = &Errors{Code: InternalServerError, Message: "could not generate a unique order number"}
	ErrOrderItemsRequired               = &Errors{Code: BadRequest, Message: "at least one order item is required"}
	ErrInvalidOrderStatusTransition     = &Errors{Code: BadRequest, Message: "invalid order status transition"}
	ErrOrderPaymentRequired             = &Errors{Code: BadRequest, Message: "order payment must be completed before accepting"}
	ErrBranchAdminMissingMerchant       = &Errors{Code: BadRequest, Message: "account must be linked to a restaurant (merchant). Sign out and sign in again, or contact support."}
	ErrRoleNotAllowedForCreation        = &Errors{Code: BadRequest, Message: "you are not allowed to create a user with this role"}
	ErrTableNotFound                    = &Errors{Code: NotFound, Message: "table not found"}
	ErrTableAlreadyExists               = &Errors{Code: BadRequest, Message: "table with this information already exists"}
	ErrReferenceNotValid                = &Errors{Code: BadRequest, Message: "reference is not valid"}
	ErrClientSessionNotFound            = &Errors{Code: NotFound, Message: "client session not found"}
	ErrClientSessionExpired             = &Errors{Code: Unauthorized, Message: "client session has expired"}
	ErrNotFound                         = &Errors{Code: NotFound, Message: "resource not found"}
	ErrRoomNotFound                     = &Errors{Code: NotFound, Message: "room not found"}
	ErrRoomAlreadyExists                = &Errors{Code: BadRequest, Message: "room with this information already exists"}
	ErrRoomAlreadyDeleted               = &Errors{Code: BadRequest, Message: "room is already deleted"}
	ErrRoomCloneAlreadyExists           = &Errors{Code: BadRequest, Message: "room type already cloned for this branch"}
	ErrRoomsHotelOnly                   = &Errors{Code: BadRequest, Message: "room types are only available for hotel merchants"}
	ErrRoomClonePriceOnly               = &Errors{Code: BadRequest, Message: "cloned room types can only customize price per night"}
	ErrRoomNumberExists                 = &Errors{Code: BadRequest, Message: "a room with this number already exists in the branch"}
	ErrRoomOccupied                     = &Errors{Code: BadRequest, Message: "this room already has an active booking"}
	ErrRoomNotOccupied                  = &Errors{Code: BadRequest, Message: "this room has no active booking"}
	ErrBookingNotFound                  = &Errors{Code: NotFound, Message: "booking not found"}
	ErrBookingNotActive                 = &Errors{Code: BadRequest, Message: "booking is not active"}
	ErrPasscodeLocked                   = &Errors{Code: TooManyRequests, Message: "too many incorrect passcode attempts; please try again later"}
	ErrInvalidPasscode                  = &Errors{Code: Unauthorized, Message: "incorrect passcode"}
	ErrRoomSessionNotFound              = &Errors{Code: NotFound, Message: "room session not found"}
	ErrRoomSessionExpired               = &Errors{Code: Unauthorized, Message: "room session has expired"}
	ErrBillNotFound                     = &Errors{Code: NotFound, Message: "room bill not found"}
	ErrBillAlreadySettled               = &Errors{Code: BadRequest, Message: "room bill is already settled"}
	ErrBillUnsettled                    = &Errors{Code: BadRequest, Message: "room bill must be settled before checkout"}
)
