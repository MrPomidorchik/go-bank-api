package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"bank-api/internal/apperror"
)

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, ErrorResponse{
		Error: ErrorBody{
			Code:    "error",
			Message: message,
		},
	})
}

func writeAppError(w http.ResponseWriter, err error) {

	if appErr, ok := errors.AsType[*apperror.AppError](err); ok {
		writeJSON(w, appErr.StatusCode, ErrorResponse{
			Error: ErrorBody{
				Code:    appErr.Code,
				Message: appErr.Message,
			},
		})
		return
	}

	writeJSON(w, http.StatusInternalServerError, ErrorResponse{
		Error: ErrorBody{
			Code:    "internal_error",
			Message: "internal server error",
		},
	})
}
