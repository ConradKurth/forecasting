package response

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Error code constants
const (
	ErrorCodeBadRequest    = "BAD_REQUEST"
	ErrorCodeUnauthorized  = "UNAUTHORIZED"
	ErrorCodeNotFound      = "NOT_FOUND"
	ErrorCodeInternalError = "INTERNAL_ERROR"
	ErrorCodeMissingParam  = "MISSING_PARAM"
	ErrorCodeInvalidToken  = "INVALID_TOKEN"
	ErrorCodeUserNotFound  = "USER_NOT_FOUND"
	ErrorCodeUnknownError  = "UNKNOWN_ERROR"
	ErrorCodeUnAuthorized  = "UNAUTHORIZED"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// HTTPError represents an error that can be returned from handlers
type HTTPError struct {
	StatusCode int
	Message    string
	ErrorCode  string
	Err        error
}

func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// NewHTTPError creates a new HTTP error
func NewHTTPError(statusCode int, errorCode, message string, err error) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		ErrorCode:  errorCode,
		Message:    message,
		Err:        err,
	}
}

// Common error constructors
func BadRequest(message string, err error) *HTTPError {
	return NewHTTPError(http.StatusBadRequest, ErrorCodeBadRequest, message, err)
}

func Unauthorized(message string, err error) *HTTPError {
	return NewHTTPError(http.StatusUnauthorized, ErrorCodeUnauthorized, message, err)
}

func NotFound(message string, err error) *HTTPError {
	return NewHTTPError(http.StatusNotFound, ErrorCodeNotFound, message, err)
}

func InternalServerError(message string, err error) *HTTPError {
	return NewHTTPError(http.StatusInternalServerError, ErrorCodeInternalError, message, err)
}

// Specific error constructors with custom codes
func MissingParameter(paramName string) *HTTPError {
	return NewHTTPError(http.StatusBadRequest, ErrorCodeMissingParam, fmt.Sprintf("Missing required parameter: %s", paramName), nil)
}

func InvalidToken() *HTTPError {
	return NewHTTPError(http.StatusUnauthorized, ErrorCodeInvalidToken, "Authentication token is invalid or expired", nil)
}

func DatabaseError(err error) *HTTPError {
	// TODO: Log the actual database error internally for debugging
	// For now, we don't expose the actual database error to the client for security reasons
	return NewHTTPError(http.StatusInternalServerError, ErrorCodeInternalError, "Internal server error", nil)
}

func UserNotFound() *HTTPError {
	return NewHTTPError(http.StatusNotFound, ErrorCodeUserNotFound, "User not found", nil)
}

// HandlerFunc represents a handler that can return an error
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// Wrap wraps a HandlerFunc to handle errors and return JSON responses
func Wrap(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			var httpErr *HTTPError

			// Check if it's already an HTTPError
			if e, ok := err.(*HTTPError); ok {
				httpErr = e
			} else {
				// Default to internal server error for unknown errors
				httpErr = InternalServerError("Internal server error", err)
			}

			// Ensure we have an error code
			if httpErr.ErrorCode == "" {
				httpErr.ErrorCode = ErrorCodeUnknownError
			}

			// Set content type to JSON
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(httpErr.StatusCode)

			// Create error response
			errorResp := ErrorResponse{
				Error:   http.StatusText(httpErr.StatusCode),
				Message: httpErr.Message,
				Code:    httpErr.ErrorCode,
			}

			// Encode and send the error response
			json.NewEncoder(w).Encode(errorResp)
		}
	}
}

// JSON writes a JSON response
func JSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}
