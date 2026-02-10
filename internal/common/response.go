package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func WriteSuccessResponse(w http.ResponseWriter, data Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(data.StatusCode)
	json.NewEncoder(w).Encode(data)

}

func WriteErrorResponse(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		var customerError *Errors
		statusCode := http.StatusInternalServerError
		message := "Internal Server Error"

		// Check if it's a custom error type
		if errors.As(err, &customerError) {
			statusCode = customerError.Code
			message = customerError.Message
		} else if validationErrors, ok := err.(validation.Errors); ok {
			// Handle ozzo-validation errors
			statusCode = http.StatusBadRequest
			var errorMessages []string
			for field, fieldError := range validationErrors {
				if fieldErr, ok := fieldError.(validation.Error); ok {
					errorMessages = append(errorMessages, fmt.Sprintf("%s: %s", field, fieldErr.Error()))
				} else {
					errorMessages = append(errorMessages, fmt.Sprintf("%s: %s", field, fieldError.Error()))
				}
			}
			message = strings.Join(errorMessages, "; ")
		} else {
			// For other errors, check if it's a validation error by checking the error message
			// or use the error message directly for better debugging
			errMsg := err.Error()
			if strings.Contains(errMsg, "validation") || strings.Contains(errMsg, "required") || strings.Contains(errMsg, "invalid") {
				statusCode = http.StatusBadRequest
				message = errMsg
			}
		}

		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(Response{
			Data:       nil,
			Message:    message,
			StatusCode: statusCode,
		})

		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(Response{
		Data:       nil,
		Message:    "Internal Server Error",
		StatusCode: http.StatusInternalServerError,
	})
}
