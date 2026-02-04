package common

import (
	"encoding/json"
	"errors"
	"net/http"
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

		if errors.As(err, &customerError) {
			statusCode = customerError.Code
			message = customerError.Message
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
