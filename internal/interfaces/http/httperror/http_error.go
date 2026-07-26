package httperror

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	appErr "sipon-api/internal/app/apperror"
	"sipon-api/internal/interfaces/http/respond"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type httpError struct {
	statusCode int
	errorCode  string
	message    string
	details    any
}

func (e *httpError) Error() string {
	return fmt.Sprintf("status: %d, code: %s, message: %s", e.statusCode, e.errorCode, e.message)
}

func badRequest(message string) *httpError {
	return &httpError{statusCode: http.StatusBadRequest, errorCode: "ERR_BAD_REQUEST", message: message}
}

func internalError(message string) *httpError {
	return &httpError{statusCode: http.StatusInternalServerError, errorCode: "ERR_INTERNAL", message: message}
}

func Handle(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// ValidationErrors ditangani di sini karena butuh payload field-by-field.
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		c.Set("request_err", err)
		c.Set("request_err_code", string(appErr.CodeUnprocessable))
		c.Set("request_err_message", string(appErr.CodeUnprocessable))
		payload := ParseValidationErrors(validationErrors)
		respond.Error(c, http.StatusUnprocessableEntity, string(appErr.CodeUnprocessable), payload)
		c.Abort()
		return
	}

	httpErr := mapError(err)
	c.Set("request_err", err)
	c.Set("request_err_code", string(httpErr.errorCode))
	c.Set("request_err_message", httpErr.message)

	if httpErr.statusCode >= 500 {
		c.Set("internal_err", err)
	}

	payload := any(httpErr.message)
	if httpErr.details != nil {
		payload = httpErr.details
	}
	respond.Error(c, httpErr.statusCode, string(httpErr.errorCode), payload)
	c.Abort()
}

func mapError(err error) *httpError {
	var appError *appErr.AppError
	if errors.As(err, &appError) {
		return mapAppError(appError)
	}

	if strings.Contains(err.Error(), "no multipart boundary") {
		return badRequest("invalid or missing multipart boundary in Content-Type header")
	}

	if errors.Is(err, io.EOF) || strings.Contains(err.Error(), "EOF") {
		return badRequest("request body is empty or not provided")
	}

	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		return badRequest(fmt.Sprintf("invalid type for field '%s'", typeErr.Field))
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return badRequest("invalid JSON format")
	}

	return internalError(string(appErr.CodeInternal))
}

func mapAppError(err *appErr.AppError) *httpError {
	switch err.Code {
	case appErr.CodeUnprocessable:
		return &httpError{statusCode: http.StatusUnprocessableEntity, errorCode: string(err.Code), message: err.Message, details: err.Details}
	case appErr.CodeUnauthorized:
		return &httpError{statusCode: http.StatusUnauthorized, errorCode: string(err.Code), message: err.Message}
	case appErr.CodeForbidden:
		return &httpError{statusCode: http.StatusForbidden, errorCode: string(err.Code), message: err.Message}
	case appErr.CodeNotFound:
		return &httpError{statusCode: http.StatusNotFound, errorCode: string(err.Code), message: err.Message}
	case appErr.CodeConflict:
		return &httpError{statusCode: http.StatusConflict, errorCode: string(err.Code), message: err.Message, details: err.Details}
	case appErr.CodeGone:
		return &httpError{statusCode: http.StatusGone, errorCode: string(err.Code), message: err.Message}
	case appErr.CodeTooManyRequests:
		return &httpError{statusCode: http.StatusTooManyRequests, errorCode: string(err.Code), message: err.Message}
	case appErr.CodeUpstreamUnavailable:
		return &httpError{statusCode: http.StatusBadGateway, errorCode: string(err.Code), message: err.Message}
	case appErr.CodeUnsupportedCurrencyPair, appErr.CodeBadRequest:
		return &httpError{statusCode: http.StatusBadRequest, errorCode: string(err.Code), message: err.Message, details: err.Details}
	default:
		return &httpError{statusCode: http.StatusInternalServerError, errorCode: string(appErr.CodeInternal), message: err.Message}
	}
}
