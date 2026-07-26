package apperror

import "fmt"

type Code string

const (
	CodeUnprocessable           Code = "ERR_UNPROCESSABLE_ENTITY"
	CodeUnauthorized            Code = "ERR_UNAUTHORIZED"
	CodeForbidden               Code = "ERR_FORBIDDEN"
	CodeNotFound                Code = "ERR_NOT_FOUND"
	CodeRouteNotFound           Code = "ERR_ROUTE_NOT_FOUND"
	CodeConflict                Code = "ERR_CONFLICT"
	CodeGone                    Code = "ERR_GONE"
	CodeInternal                Code = "ERR_INTERNAL"
	CodeTooManyRequests         Code = "ERR_TOO_MANY_REQUESTS"
	CodeUpstreamUnavailable     Code = "ERR_UPSTREAM_UNAVAILABLE"
	CodeUnsupportedCurrencyPair Code = "ERR_UNSUPPORTED_CURRENCY_PAIR"
	CodeBadRequest              Code = "ERR_BAD_REQUEST"
)

// Validation codes — cross-domain technical concerns
const (
	CodeInvalidMediaKey           Code = "INVALID_MEDIA_KEY"
	CodeContentTypeNotAllowed     Code = "CONTENT_TYPE_NOT_ALLOWED"
	CodeInvalidDateFormat         Code = "INVALID_DATE_FORMAT"
	CodeSearchQueryRequired       Code = "SEARCH_QUERY_REQUIRED"
	CodeTokenRequired             Code = "TOKEN_REQUIRED"
	CodePasswordSameAsCurrent     Code = "PASSWORD_SAME_AS_CURRENT"
	CodeUnsupportedIdentifierKind Code = "UNSUPPORTED_IDENTIFIER_KIND"
	CodeContentBannedKeyword      Code = "CONTENT_BANNED_KEYWORD"
)

type AppError struct {
	Code    Code
	Message string
	Details any
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("code: %s, message: %s, err: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("code: %s, message: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code Code, message string, details any) *AppError {
	return &AppError{Code: code, Message: message, Details: details}
}

func Wrap(code Code, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

func Unprocessable(message string, details any, cause ...error) *AppError {
	ae := New(CodeUnprocessable, message, details)
	ae.Err = firstCause(cause...)
	return ae
}

func Unauthorized(message string, cause ...error) *AppError {
	ae := New(CodeUnauthorized, message, nil)
	ae.Err = firstCause(cause...)
	return ae
}

func Forbidden(message string, cause ...error) *AppError {
	ae := New(CodeForbidden, message, nil)
	ae.Err = firstCause(cause...)
	return ae
}

func NotFound(message string, cause ...error) *AppError {
	ae := New(CodeNotFound, message, nil)
	ae.Err = firstCause(cause...)
	return ae
}

func Conflict(message string, cause ...error) *AppError {
	ae := New(CodeConflict, message, nil)
	ae.Err = firstCause(cause...)
	return ae
}

// ConflictWithDetails sama seperti Conflict tapi menyertakan payload data
// terstruktur (mis. existing_request_id) di response body.
func ConflictWithDetails(message string, details any) *AppError {
	return New(CodeConflict, message, details)
}

// Gone menandakan resource pernah ada tapi sudah tidak berlaku (kedaluwarsa/dicabut/kuota habis),
// dibedakan dari NotFound agar client bisa memberi pesan yang tepat (410 vs 404).
func Gone(message string, cause ...error) *AppError {
	ae := New(CodeGone, message, nil)
	ae.Err = firstCause(cause...)
	return ae
}

func BadRequest(message string, cause ...error) *AppError {
	ae := New(CodeBadRequest, message, nil)
	ae.Err = firstCause(cause...)
	return ae
}

// BadRequestWithDetails sama seperti BadRequest tapi menyertakan payload data
// terstruktur (mis. payment_provider) di response body.
func BadRequestWithDetails(message string, details any) *AppError {
	return New(CodeBadRequest, message, details)
}

func Internal(message string, err error) *AppError {
	return Wrap(CodeInternal, message, err)
}

func TooManyRequests(message string, cause ...error) *AppError {
	ae := New(CodeTooManyRequests, message, nil)
	ae.Err = firstCause(cause...)
	return ae
}

func UpstreamUnavailable(message string, err error) *AppError {
	return Wrap(CodeUpstreamUnavailable, message, err)
}

func firstCause(causes ...error) error {
	for _, err := range causes {
		if err != nil {
			return err
		}
	}
	return nil
}
